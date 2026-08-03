package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/contracts"
)

// scriptTimeout caps how long a skill script may run. Image generation
// (doubao ~20s + OSS upload ~60s) fits well under this.
const scriptTimeout = 120 * time.Second

// maxScriptOutput caps returned stdout/stderr to avoid blowing up the LLM context.
const maxScriptOutput = 20000

// interpreters maps an allowed script extension to its interpreter binary.
// Only these extensions are permitted — no shell, no arbitrary interpreters.
var interpreters = map[string]string{
	".py": "python3",
	".sh": "sh",
}

// RunSkillScriptTool executes a script bundled inside a skill package.
// The script is streamed from the skill's OSS zip into a per-call temp dir and
// executed directly via argv (never through a shell), so prompt arguments are
// passed safely without injection risk. Only skills granted to the agent may
// run scripts.
type RunSkillScriptTool struct {
	loader contracts.SkillsLoaderInterface
}

func NewRunSkillScriptTool(loader contracts.SkillsLoaderInterface) tool.InvokableTool {
	return &RunSkillScriptTool{loader: loader}
}

func (t *RunSkillScriptTool) Name() string { return "run_skill_script" }

func (t *RunSkillScriptTool) Description() string {
	return "Run a script bundled inside a skill package and return its stdout. " +
		"The skill must be in the available-skills list. " +
		"script_path is relative to the skill root (e.g. \"scripts/image_client.py\"). " +
		"args is a pre-split argv list (do NOT shell-quote); pass each argument as a separate array element."
}

func (t *RunSkillScriptTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"skill_name": {
				Type:     schema.String,
				Desc:     "The name of the skill (from the available-skills list in system prompt)",
				Required: true,
			},
			"script_path": {
				Type:     schema.String,
				Desc:     "Path of the script inside the skill, relative to the skill root (e.g. \"scripts/image_client.py\")",
				Required: true,
			},
			"args": {
				Type:     schema.Array,
				Desc:     "Pre-split argv passed to the script. Each element is one argument; do not shell-quote. Optional.",
				Required: false,
			},
		}),
	}, nil
}

type runSkillScriptInput struct {
	SkillName  string   `json:"skill_name"`
	ScriptPath string   `json:"script_path"`
	Args       []string `json:"args"`
}

func (t *RunSkillScriptTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input runSkillScriptInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.SkillName == "" || input.ScriptPath == "" {
		return "", fmt.Errorf("skill_name and script_path are required")
	}

	// 1. extension whitelist → interpreter
	ext := strings.ToLower(filepath.Ext(input.ScriptPath))
	interpreter, ok := interpreters[ext]
	if !ok {
		return "", fmt.Errorf("unsupported script type %q (allowed: .py, .sh)", ext)
	}

	// 2. interpreter present on the host?
	if _, err := exec.LookPath(interpreter); err != nil {
		return "", fmt.Errorf("interpreter %q not found on server: %w", interpreter, err)
	}

	// 3. authorization: skill must be granted to this agent (has side effects)
	if !t.skillGranted(input.SkillName) {
		return "", fmt.Errorf("skill '%s' is not granted to this agent", input.SkillName)
	}

	// 4. fetch script bytes from OSS zip
	data, found, err := t.loader.GetSkillFile(input.SkillName, input.ScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to fetch skill script: %w", err)
	}
	if !found {
		return "", fmt.Errorf("script '%s' not found in skill '%s'", input.ScriptPath, input.SkillName)
	}

	// 5. write to a per-call temp dir, clean up when done
	tmpDir, err := os.MkdirTemp("", "skillscript-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptFile := filepath.Join(tmpDir, filepath.Base(input.ScriptPath))
	if err := os.WriteFile(scriptFile, data, 0700); err != nil {
		return "", fmt.Errorf("failed to write script to temp: %w", err)
	}

	// 6. execute directly via argv (no shell), scoped to temp dir, with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()

	args := append([]string{scriptFile}, input.Args...)
	cmd := exec.CommandContext(cmdCtx, interpreter, args...)
	cmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("script timed out after %s", scriptTimeout)
	}
	if runErr != nil {
		output += fmt.Sprintf("\nExit code: %v", runErr)
	}

	if output == "" {
		output = "(no output)"
	}
	if len(output) > maxScriptOutput {
		output = output[:maxScriptOutput] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxScriptOutput)
	}
	return output, nil
}

// skillGranted reports whether the skill is in the agent's granted skill list.
// ListSkills already filters by agent grants, so membership = authorized.
func (t *RunSkillScriptTool) skillGranted(name string) bool {
	for _, s := range t.loader.ListSkills() {
		if s.Name == name {
			return true
		}
	}
	return false
}

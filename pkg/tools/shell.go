package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

var (
	defaultDenyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),
		regexp.MustCompile(`\bdel\s+/[fq]\b`),
		regexp.MustCompile(`\brmdir\s+/s\b`),
		regexp.MustCompile(`\b(format|mkfs|diskpart)\b\s`),
		regexp.MustCompile(`\bdd\s+if=`),
		regexp.MustCompile(`>\s*/dev/sd[a-z]\b`),
		regexp.MustCompile(`\b(shutdown|reboot|poweroff)\b`),
		regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
		regexp.MustCompile(`\$\([^)]+\)`),
		regexp.MustCompile(`\$\{[^}]+\}`),
		regexp.MustCompile("`[^`]+`"),
		regexp.MustCompile(`\|\s*sh\b`),
		regexp.MustCompile(`\|\s*bash\b`),
		regexp.MustCompile(`;\s*rm\s+-[rf]`),
		regexp.MustCompile(`&&\s*rm\s+-[rf]`),
		regexp.MustCompile(`\|\|\s*rm\s+-[rf]`),
		regexp.MustCompile(`<<\s*EOF`),
		regexp.MustCompile(`\$\(\s*cat\s+`),
		regexp.MustCompile(`\$\(\s*curl\s+`),
		regexp.MustCompile(`\$\(\s*wget\s+`),
		regexp.MustCompile(`\$\(\s*which\s+`),
		regexp.MustCompile(`\bsudo\b`),
		regexp.MustCompile(`\bchmod\s+[0-7]{3,4}\b`),
		regexp.MustCompile(`\bchown\b`),
		regexp.MustCompile(`\bpkill\b`),
		regexp.MustCompile(`\bkillall\b`),
		regexp.MustCompile(`\bkill\b`),
		regexp.MustCompile(`\bcurl\b.*\|\s*(sh|bash)`),
		regexp.MustCompile(`\bwget\b.*\|\s*(sh|bash)`),
		regexp.MustCompile(`\bnpm\s+install\s+-g\b`),
		regexp.MustCompile(`\bpip\s+install\s+--user\b`),
		regexp.MustCompile(`\bapt\s+(install|remove|purge)\b`),
		regexp.MustCompile(`\byum\s+(install|remove)\b`),
		regexp.MustCompile(`\bdnf\s+(install|remove)\b`),
		regexp.MustCompile(`\bdocker\s+run\b`),
		regexp.MustCompile(`\bdocker\s+exec\b`),
		regexp.MustCompile(`\bgit\s+push\b`),
		regexp.MustCompile(`\bgit\s+force\b`),
		regexp.MustCompile(`\bssh\b.*@`),
		regexp.MustCompile(`\beval\b`),
		regexp.MustCompile(`\bsource\s+.*\.sh\b`),
		regexp.MustCompile(`\bpython[23]?\s+-c\b`),
		regexp.MustCompile(`\bperl\s+-e\b`),
		regexp.MustCompile(`\bruby\s+-e\b`),
		regexp.MustCompile(`\bnode\s+-e\b`),
		regexp.MustCompile(`\bnc\b`),
		regexp.MustCompile(`\bnetcat\b`),
		regexp.MustCompile(`\bncat\b`),
		regexp.MustCompile(`\blua\s+-e\b`),
		regexp.MustCompile(`\bphp\s+-r\b`),
		regexp.MustCompile(`/etc/shadow`),
	}

	absolutePathPattern = regexp.MustCompile(`[A-Za-z]:\\[^\\\"']+|/[^\s\"']+`)

	safePaths = map[string]bool{
		"/dev/null":    true,
		"/dev/zero":    true,
		"/dev/random":  true,
		"/dev/urandom": true,
		"/dev/stdin":   true,
		"/dev/stdout":  true,
		"/dev/stderr":  true,
	}
)

func guardCommand(command, cwd string, restrictToWorkspace bool) string {
	cmd := strings.TrimSpace(command)
	lower := strings.ToLower(cmd)

	for _, pattern := range defaultDenyPatterns {
		if pattern.MatchString(lower) {
			return "Command blocked by safety guard (dangerous pattern detected)"
		}
	}

	if restrictToWorkspace {
		if strings.Contains(cmd, "..\\") || strings.Contains(cmd, "../") {
			return "Command blocked by safety guard (path traversal detected)"
		}

		cwdPath, err := filepath.Abs(cwd)
		if err != nil {
			return ""
		}

		matches := absolutePathPattern.FindAllString(cmd, -1)

		for _, raw := range matches {
			p, err := filepath.Abs(raw)
			if err != nil {
				continue
			}

			if safePaths[p] {
				continue
			}

			rel, err := filepath.Rel(cwdPath, p)
			if err != nil {
				continue
			}

			if strings.HasPrefix(rel, "..") {
				return "Command blocked by safety guard (path outside working dir)"
			}
		}
	}

	return ""
}

type ExecInput struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir,omitempty"`
}

type ExecOutput struct {
	Output string `json:"output"`
}

func NewExecTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[ExecInput, ExecOutput](
		&schema.ToolInfo{
			Name: "exec",
			Desc: "Execute a shell command and return its output. Use with caution.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"command": {
					Type:     schema.String,
					Desc:     "The shell command to execute",
					Required: true,
				},
				"working_dir": {
					Type: schema.String,
					Desc: "Optional working directory for the command",
				},
			}),
		},
		func(ctx context.Context, input ExecInput) (ExecOutput, error) {
			if input.Command == "" {
				return ExecOutput{}, fmt.Errorf("command is required")
			}

			cwd := WorkspaceFromContext(ctx)
			if input.WorkingDir != "" {
				cwd = input.WorkingDir
			}

			if cwd == "" {
				if wd, err := os.Getwd(); err == nil {
					cwd = wd
				}
			}

			if guardErr := guardCommand(input.Command, cwd, restrict); guardErr != "" {
				return ExecOutput{}, fmt.Errorf("%s", guardErr)
			}

			cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(cmdCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", input.Command)
			} else {
				cmd = exec.CommandContext(cmdCtx, "sh", "-c", input.Command)
			}
			if cwd != "" {
				cmd.Dir = cwd
			}

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String()
			if stderr.Len() > 0 {
				output += "\nSTDERR:\n" + stderr.String()
			}

			if err != nil {
				if cmdCtx.Err() == context.DeadlineExceeded {
					return ExecOutput{}, fmt.Errorf("command timed out")
				}
				output += fmt.Sprintf("\nExit code: %v", err)
			}

			if output == "" {
				output = "(no output)"
			}

			maxLen := 10000
			if len(output) > maxLen {
				output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
			}

			return ExecOutput{Output: output}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

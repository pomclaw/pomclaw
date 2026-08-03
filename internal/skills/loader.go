package skills

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/oss"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// skillZipCacheTTL 是 skill zip 包在 Redis 中的缓存时长。版本号已包含在
// cache key 中，技能更新（version 自增）会自动失效；TTL 仅作为同版本重传
// 的兜底与闲置条目淘汰。
const skillZipCacheTTL = 24 * 60 * 60

type DBSkillsLoader struct {
	skillsModel      model.SkillsModel
	skillGrantsModel model.SkillGrantsModel
	ossClient        *oss.Client
	redisClient      *redis.Redis
	agentID          string
	userID           string
}

func NewDBSkillsLoader(
	skillsModel model.SkillsModel,
	skillGrantsModel model.SkillGrantsModel,
	ossClient *oss.Client,
	redisClient *redis.Redis,
	agentID string,
	userID string,
) contracts.SkillsLoaderInterface {
	return &DBSkillsLoader{
		skillsModel:      skillsModel,
		skillGrantsModel: skillGrantsModel,
		ossClient:        ossClient,
		redisClient:      redisClient,
		agentID:          agentID,
		userID:           userID,
	}
}

func (l *DBSkillsLoader) ListSkills() []contracts.SkillInfo {
	ctx := context.Background()

	grants, err := l.skillGrantsModel.FindByAgentID(ctx, l.agentID)
	if err != nil || len(grants) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(grants))
	for _, g := range grants {
		ids = append(ids, g.SkillId)
	}

	skills, err := l.skillsModel.FindEnabledByIDs(ctx, ids)
	if err != nil {
		return nil
	}

	result := make([]contracts.SkillInfo, 0, len(skills))
	for _, s := range skills {
		desc := ""
		if s.Description.Valid {
			desc = s.Description.String
		}
		result = append(result, contracts.SkillInfo{
			Name:        s.Name,
			Path:        fmt.Sprintf("%s/%s-%d.zip", l.userID, s.Slug, s.Version),
			Source:      "oss",
			Description: desc,
		})
	}
	return result
}

func (l *DBSkillsLoader) LoadSkill(name string) (string, bool) {
	ctx := context.Background()

	skill, err := l.skillsModel.FindByName(ctx, l.userID, name)
	if err != nil {
		return "", false
	}

	if l.ossClient == nil {
		return "", false
	}

	data, err := l.getSkillZip(ctx, skill.Slug, int(skill.Version))
	if err != nil {
		return "", false
	}

	content, err := extractSkillContent(data)
	if err != nil {
		return "", false
	}

	return content, true
}

// getSkillZip 读取一个 skill 的整包 zip 字节，优先走 Redis 缓存。
// cache key 含 userID + slug + version，技能更新（version 自增）自动失效。
// Redis 未配置或出错时降级直连 OSS，不影响功能。只缓存成功的 zip 字节。
func (l *DBSkillsLoader) getSkillZip(ctx context.Context, slug string, version int) ([]byte, error) {
	objectKey := l.ossClient.BuildObjectKey(l.userID, slug, version)

	if l.redisClient != nil {
		key := fmt.Sprintf("skill:zip:%s:%s:%d", l.userID, slug, version)
		if cached, err := l.redisClient.GetCtx(ctx, key); err != nil {
			logx.Errorf("redis get skill zip failed, fallback to oss: %v", err)
		} else if cached != "" {
			return []byte(cached), nil
		}
	}

	data, err := l.ossClient.GetContent(ctx, objectKey)
	if err != nil {
		return nil, err
	}

	if l.redisClient != nil {
		key := fmt.Sprintf("skill:zip:%s:%s:%d", l.userID, slug, version)
		// 尽力写回缓存，失败不影响本次返回
		if err := l.redisClient.SetexCtx(ctx, key, string(data), skillZipCacheTTL); err != nil {
			logx.Errorf("redis setex skill zip failed: %v", err)
		}
	}

	return data, nil
}

// GetSkillFile reads a single file from a skill's OSS zip by relative path.
// See contracts.SkillsLoaderInterface.GetSkillFile for semantics.
func (l *DBSkillsLoader) GetSkillFile(name string, filePath string) ([]byte, bool, error) {
	ctx := context.Background()

	skill, err := l.skillsModel.FindByName(ctx, l.userID, name)
	if err != nil {
		// skill not found → (nil, false, nil), consistent with LoadSkill
		return nil, false, nil
	}

	if l.ossClient == nil {
		return nil, false, nil
	}

	data, err := l.getSkillZip(ctx, skill.Slug, int(skill.Version))
	if err != nil {
		return nil, false, err
	}

	fileData, found, err := extractFile(data, filePath)
	if err != nil {
		return nil, false, err
	}
	if !found {
		// file not found within zip → (nil, false, nil)
		return nil, false, nil
	}

	return fileData, true, nil
}

func (l *DBSkillsLoader) BuildSkillsSummary() string {
	skills := l.ListSkills()
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<available-skills>\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("  <skill name=\"%s\" description=\"%s\" />\n", s.Name, s.Description))
	}
	sb.WriteString("</available-skills>")
	return sb.String()
}

func extractSkillContent(zipData []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", err
	}

	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		base := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			base = name[idx+1:]
		}
		if base == "skill.md" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			return stripFrontmatter(string(content)), nil
		}
	}
	return "", fmt.Errorf("SKILL.md not found")
}

// extractFile reads a single file from a skill zip by relative path.
// targetPath is relative to the skill root (e.g. "references/image-rules.md").
// Entries are matched after stripping an optional single top-level directory
// prefix, with a fallback to the full path. Matching is case-insensitive and
// slash-normalized. Returns (bytes, true, nil) on hit, (nil, false, nil) when
// no entry matches, and (nil, false, err) only on zip-read errors.
func extractFile(zipData []byte, targetPath string) ([]byte, bool, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, false, err
	}

	target := normalizeRelPath(targetPath)

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if normalizeRelPath(stripTopDir(f.Name)) == target {
			rc, err := f.Open()
			if err != nil {
				return nil, false, err
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, false, err
			}
			return data, true, nil
		}
	}

	// fallback: match without stripping top-level dir (covers multi-segment prefixes)
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if normalizeRelPath(f.Name) == target {
			rc, err := f.Open()
			if err != nil {
				return nil, false, err
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, false, err
			}
			return data, true, nil
		}
	}

	return nil, false, nil
}

// stripTopDir removes a single leading directory segment from a zip entry path.
// "dragon-journey/scripts/x.py" → "scripts/x.py"; "scripts/x.py" → "scripts/x.py".
func stripTopDir(name string) string {
	name = strings.TrimPrefix(name, "./")
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		return strings.Join(parts[1:], "/")
	}
	return name
}

// normalizeRelPath normalizes a relative path for case-insensitive comparison:
// trims whitespace, strips "./", converts backslashes to forward slashes, lowercases.
func normalizeRelPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.ReplaceAll(p, "\\", "/")
	return strings.ToLower(p)
}

func stripFrontmatter(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return content
	}
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return content
	}
	return strings.TrimSpace(content[end+6:])
}

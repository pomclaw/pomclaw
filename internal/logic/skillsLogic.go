package logic

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

type SkillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SkillsLogic {
	return &SkillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// --- helpers ---

// getOwnedSkill 校验 skill 归属当前用户
func (l *SkillsLogic) getOwnedSkill(id int64) (*model.Skills, error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}
	skill, err := l.svcCtx.SkillsModel.FindOne(l.ctx, id)
	if errors.Is(err, model.ErrNotFound) || (err == nil && skill.UserId != userID) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return skill, nil
}

func (l *SkillsLogic) toSkillResp(s *model.Skills) types.SkillResp {
	return types.SkillResp{
		ID:          s.Id,
		Name:        s.Name,
		Slug:        s.Slug,
		Description: nullStringToString(s.Description),
		Enabled:     s.Enabled,
		Status:      s.Status,
		Version:     int(s.Version),
		IsSystem:    false,
		Source:      "file",
		Visibility:  "private",
		IsShared:    s.IsShared,
		CreatedBy:   s.UserId,
	}
}

// getSkillZipReader 从 OSS 下载 zip 并返回 reader
func (l *SkillsLogic) getSkillZipReader(skill *model.Skills, version int) (*zip.Reader, error) {
	if l.svcCtx.OssClient == nil {
		return nil, fmt.Errorf("OSS not configured")
	}
	if version <= 0 {
		version = int(skill.Version)
	}
	objectKey := l.svcCtx.OssClient.BuildObjectKey(skill.UserId, skill.Slug, version)
	data, err := l.svcCtx.OssClient.GetContent(l.ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("fetch skill zip: %w", err)
	}
	return zip.NewReader(bytes.NewReader(data), int64(len(data)))
}

// readSkillContent 从 zip 中提取 SKILL.md 或 README.md 的文本内容
func (l *SkillsLogic) readSkillContent(skill *model.Skills) string {
	zipReader, err := l.getSkillZipReader(skill, 0)
	if err != nil {
		l.Infof("[skills] readSkillContent: skip, zip open failed for skill=%d: %v", skill.Id, err)
		return ""
	}
	for _, f := range zipReader.File {
		name := strings.ToLower(filepath.Base(f.Name))
		if name == "skill.md" || name == "readme.md" {
			rc, err := f.Open()
			if err != nil {
				return ""
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return ""
			}
			return string(content)
		}
	}
	return ""
}

// --- public methods ---

// GetSkill 获取 skill 详情（含 content）
func (l *SkillsLogic) GetSkill(req *types.GetSkillReq) (*types.GetSkillResp, error) {
	skill, err := l.getOwnedSkill(req.ID)
	if err != nil {
		return nil, err
	}
	l.Infof("[skills] GetSkill: id=%d slug=%s", skill.Id, skill.Slug)
	resp := l.toSkillResp(skill)
	content := l.readSkillContent(skill)
	return &types.GetSkillResp{Skill: resp, Content: content}, nil
}

// UpdateSkill 更新 skill 属性；status="deleted" 触发软删除
func (l *SkillsLogic) UpdateSkill(req *types.UpdateSkillReq) (*types.UpdateSkillResp, error) {
	skill, err := l.getOwnedSkill(req.ID)
	if err != nil {
		return nil, err
	}

	skill.Enabled = req.Enabled
	skill.IsShared = req.IsShared
	if req.Name != "" {
		skill.Name = req.Name
	}
	if req.Description != "" {
		skill.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != "" {
		skill.Status = req.Status
	}
	if req.Status == "deleted" {
		skill.Enabled = false
	}
	skill.UpdatedAt = time.Now()

	if err := l.svcCtx.SkillsModel.Update(l.ctx, skill); err != nil {
		return nil, err
	}

	l.Infof("[skills] UpdateSkill: id=%d status=%s enabled=%v", skill.Id, skill.Status, skill.Enabled)
	resp := l.toSkillResp(skill)
	return &types.UpdateSkillResp{Skill: resp}, nil
}

// ListSkills 列出当前用户可见的所有 skill（含分享的），富化 agent_count 和 created_by_name
func (l *SkillsLogic) ListSkills(req *types.ListSkillsReq) (*types.SkillsResp, error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	skills, err := l.svcCtx.SkillsModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	// Batch query grant counts
	ids := make([]int64, 0, len(skills))
	for _, s := range skills {
		ids = append(ids, s.Id)
	}
	counts, err := l.svcCtx.SkillGrantsModel.CountBySkillIDs(l.ctx, ids)
	if err != nil {
		l.Infof("[skills] ListSkills: failed to count grants: %v", err)
	}

	// Batch query usernames
	uidSet := make([]string, 0, len(skills))
	for _, s := range skills {
		if !slices.Contains(uidSet, s.UserId) {
			uidSet = append(uidSet, s.UserId)
		}
	}
	userMap := make(map[string]string, len(uidSet))
	if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, uidSet); err == nil {
		for _, u := range users {
			userMap[u.UserId] = u.Username
		}
	}

	skillList := make([]types.SkillResp, 0, len(skills))
	for _, s := range skills {
		svr := l.toSkillResp(s)
		svr.AgentCount = int(counts[s.Id])
		if name, ok := userMap[s.UserId]; ok {
			svr.CreatedByName = name
		}
		skillList = append(skillList, svr)
	}

	return &types.SkillsResp{
		Skills: skillList,
	}, nil
}

// ListAgentSkills 列出 skill 并标记该 agent 的授权状态
func (l *SkillsLogic) ListAgentSkills(req *types.ListAgentSkillsReq) (*types.SkillsWithGrantResp, error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	skills, err := l.svcCtx.SkillsModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	grants, err := l.svcCtx.SkillGrantsModel.FindByAgentID(l.ctx, req.AgentID)
	if err != nil {
		return nil, err
	}

	grantedSet := make(map[int64]bool, len(grants))
	for _, g := range grants {
		grantedSet[g.SkillId] = true
	}

	// Batch query grant counts
	ids := make([]int64, 0, len(skills))
	for _, s := range skills {
		ids = append(ids, s.Id)
	}
	counts, err := l.svcCtx.SkillGrantsModel.CountBySkillIDs(l.ctx, ids)
	if err != nil {
		l.Infof("[skills] ListAgentSkills: failed to count grants: %v", err)
	}

	// Batch query usernames
	uidSet := make([]string, 0, len(skills))
	for _, s := range skills {
		if !slices.Contains(uidSet, s.UserId) {
			uidSet = append(uidSet, s.UserId)
		}
	}
	userMap := make(map[string]string, len(uidSet))
	if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, uidSet); err == nil {
		for _, u := range users {
			userMap[u.UserId] = u.Username
		}
	}

	resp := make([]types.SkillWithGrantResp, 0, len(skills))
	for _, s := range skills {
		svr := types.SkillWithGrantResp{
			ID:          s.Id,
			Name:        s.Name,
			Slug:        s.Slug,
			Description: nullStringToString(s.Description),
			Enabled:     s.Enabled,
			Status:      s.Status,
			Version:     int(s.Version),
			IsSystem:    false,
			Source:      "file",
			Visibility:  "private",
			Granted:     grantedSet[s.Id],
			IsShared:    s.IsShared,
			CreatedBy:   s.UserId,
			AgentCount:  int(counts[s.Id]),
		}
		if name, ok := userMap[s.UserId]; ok {
			svr.CreatedByName = name
		}
		resp = append(resp, svr)
	}

	return &types.SkillsWithGrantResp{Skills: resp}, nil
}

// ListSkillGrants 列出 skill 的所有 agent 授权
func (l *SkillsLogic) ListSkillGrants(req *types.ListSkillGrantsReq) (*types.ListSkillGrantsResp, error) {
	grants, err := l.svcCtx.SkillGrantsModel.FindBySkillID(l.ctx, req.ID)
	if err != nil {
		l.Errorf("ListSkillGrants failed: %v", err)
		return nil, err
	}

	grantList := make([]types.SkillAgentGrant, 0, len(grants))
	for _, g := range grants {
		grantList = append(grantList, types.SkillAgentGrant{
			ID:        fmt.Sprintf("%d", g.Id),
			SkillID:   g.SkillId,
			AgentID:   g.AgentId,
			Enabled:   true,
			CreatedAt: g.CreatedAt.Unix(),
		})
	}

	return &types.ListSkillGrantsResp{AgentGrants: grantList}, nil
}

// GrantSkillAgent 授权 skill 给 agent
func (l *SkillsLogic) GrantSkillAgent(req *types.GrantSkillAgentReq) (*types.GrantSkillAgentResp, error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.SkillGrantsModel.InsertGrant(l.ctx, userID, req.ID, req.AgentID); err != nil {
		l.Errorf("GrantSkillAgent failed: %v", err)
		return nil, err
	}

	l.Infof("[skills] GrantSkillAgent: skill=%d agent=%s", req.ID, req.AgentID)
	return &types.GrantSkillAgentResp{
		Grant: types.SkillAgentGrant{
			SkillID: req.ID,
			AgentID: req.AgentID,
			Enabled: true,
		},
	}, nil
}

// RevokeSkillAgentGrant 撤销 skill 对 agent 的授权
func (l *SkillsLogic) RevokeSkillAgentGrant(req *types.RevokeSkillAgentGrantReq) (*types.RevokeSkillAgentGrantResp, error) {
	if err := l.svcCtx.SkillGrantsModel.DeleteBySkillIDAgentID(l.ctx, req.ID, req.AgentID); err != nil {
		l.Errorf("RevokeSkillAgentGrant failed: %v", err)
		return nil, err
	}

	l.Infof("[skills] RevokeSkillAgentGrant: skill=%d agent=%s", req.ID, req.AgentID)
	return &types.RevokeSkillAgentGrantResp{}, nil
}

// --- upload ---
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug"`
	Description string `yaml:"description"`
}

// UploadSkill 处理 zip 上传：解析 frontmatter → 计算 hash → 上传 OSS → 写入/更新 DB
func (l *SkillsLogic) UploadSkill(file multipart.File) (*types.UploadSkillResp, error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	meta, data, err := l.parseUploadedZip(file)
	if err != nil {
		return nil, err
	}

	existing, err := l.svcCtx.SkillsModel.FindOneByUserIdSlug(l.ctx, userID, meta.Slug)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	isNew := existing == nil
	version := int64(1)
	if !isNew {
		version = existing.Version + 1
	}

	contentURL, err := l.uploadToOSS(userID, meta.Slug, int(version), data)
	if err != nil {
		return nil, err
	}

	fileHash := hashSHA256(data)
	skillID, err := l.upsertSkillRecord(userID, meta, existing, version, contentURL, fileHash)
	if err != nil {
		return nil, err
	}

	l.Infof("[skills] UploadSkill: slug=%s version=%d isNew=%v", meta.Slug, version, isNew)
	return &types.UploadSkillResp{
		ID:         skillID,
		Slug:       meta.Slug,
		Version:    int(version),
		Name:       meta.Name,
		Status:     "active",
		IsNew:      isNew,
		ContentUrl: contentURL,
	}, nil
}

// parseUploadedZip 读取文件内容并从 zip 中提取 skill metadata
func (l *SkillsLogic) parseUploadedZip(file multipart.File) (*skillFrontmatter, []byte, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("read file: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid zip: %w", err)
	}

	meta, err := extractSkillMeta(zipReader)
	if err != nil {
		return nil, nil, fmt.Errorf("parse skill metadata: %w", err)
	}
	return meta, data, nil
}

// uploadToOSS 上传 zip 到对象存储，返回访问 URL
func (l *SkillsLogic) uploadToOSS(userID, slug string, version int, data []byte) (string, error) {
	if l.svcCtx.OssClient == nil {
		return "", fmt.Errorf("OSS not configured")
	}
	objectKey := l.svcCtx.OssClient.BuildObjectKey(userID, slug, version)
	contentURL, err := l.svcCtx.OssClient.UploadFile(l.ctx, objectKey, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("oss upload: %w", err)
	}
	return contentURL, nil
}

// upsertSkillRecord 新建或更新 skill 数据库记录
func (l *SkillsLogic) upsertSkillRecord(
	userID string, meta *skillFrontmatter, existing *model.Skills,
	version int64, contentURL, fileHash string,
) (int64, error) {
	if existing == nil {
		skill := &model.Skills{
			UserId:      userID,
			Name:        meta.Name,
			Slug:        meta.Slug,
			Description: sql.NullString{String: meta.Description, Valid: meta.Description != ""},
			Enabled:     true,
			Status:      "active",
			Version:     version,
		}
		id, err := l.svcCtx.SkillsModel.InsertReturningID(l.ctx, skill)
		if err != nil {
			return 0, fmt.Errorf("insert skill: %w", err)
		}
		return id, nil
	}

	existing.Name = meta.Name
	existing.Description = sql.NullString{String: meta.Description, Valid: meta.Description != ""}
	existing.Version = version
	existing.Status = "active"
	if err := l.svcCtx.SkillsModel.Update(l.ctx, existing); err != nil {
		return 0, fmt.Errorf("update skill: %w", err)
	}
	return existing.Id, nil
}

// --- frontmatter parsing ---

func extractSkillMeta(zipReader *zip.Reader) (*skillFrontmatter, error) {
	for _, f := range zipReader.File {
		name := strings.ToLower(filepath.Base(f.Name))
		if name == "skill.md" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			return parseFrontmatter(string(content))
		}
	}
	return nil, fmt.Errorf("SKILL.md not found in zip")
}

func parseFrontmatter(content string) (*skillFrontmatter, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter found")
	}
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return nil, fmt.Errorf("frontmatter not closed")
	}
	fmContent := content[3 : end+3]

	var meta skillFrontmatter
	if err := yaml.Unmarshal([]byte(fmContent), &meta); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if meta.Name == "" {
		return nil, fmt.Errorf("skill name required in frontmatter")
	}
	if meta.Slug == "" {
		meta.Slug = strings.ReplaceAll(strings.ToLower(meta.Name), " ", "-")
	}
	return &meta, nil
}

func hashSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

package logic

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyProviderLogic {
	return &VerifyProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyProviderLogic) VerifyProvider(req types.VerifyProviderReq) (*types.VerifyProviderResp, error) {
	if req.Model == "" {
		return &types.VerifyProviderResp{
			Valid: false,
			Error: "model is required",
		}, nil
	}

	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.ID)
	if err == model.ErrNotFound || (err == nil && p.UserId != req.UserID) {
		l.Errorf("VerifyProvider: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("VerifyProvider failed: %v", err)
		return nil, err
	}

	llm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  p.ApiKey,
		BaseURL: p.ApiBase.String,
		Model:   req.Model,
	})
	if err != nil {
		return nil, err
	}

	_, err = llm.Generate(l.ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: "你好",
		},
	})
	if err != nil {
		return &types.VerifyProviderResp{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &types.VerifyProviderResp{Valid: true}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"golang.org/x/crypto/bcrypt"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// User login
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.AuthResp, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	user, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Generate JWT token using go-zero's approach
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, accessExpire, user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &types.AuthResp{
		AccessToken:  accessToken,
		RefreshToken: "", // TODO: implement refresh token if needed
		ExpiresIn:    accessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (l *LoginLogic) getJwtToken(secretKey string, seconds int64, userId string) (string, error) {
	now := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = now + seconds
	claims["iat"] = now
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

package handler

import (
	"context"
	"easy-chat/apps/im/ws/internal/svc"
	"easy-chat/pkg/ctxdata"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/token"
	"net/http"
)

type JwtAuth struct {
	svc    *svc.ServiceContext
	parser *token.TokenParser
	logx.Logger
}

func NewJwtAuth(svc *svc.ServiceContext) *JwtAuth {
	return &JwtAuth{
		svc:    svc,
		parser: token.NewTokenParser(),
		Logger: logx.WithContext(context.Background()),
	}
}

func (j *JwtAuth) Auth(w http.ResponseWriter, r *http.Request) bool {
	//处理websocket子协议认证问题，由于websocket协议本身无法携带Header信息，需要使用子协议
	//这里直接去获取子协议的Header信息，为其设置token
	if tok := r.Header.Get("Sec-Websocket-Protocol"); tok != "" {
		r.Header.Set("Authorization", tok)
	}
	//解析token
	fmt.Println("接收到的token为：", j.svc.Config.JwtAuth.AccessSecret)
	tok, err := j.parser.ParseToken(r, j.svc.Config.JwtAuth.AccessSecret, "")
	if err != nil {
		j.Errorf("parse token err %v ", err)
		return false
	}

	if !tok.Valid {
		return false
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	//把token数据传递至上下文中
	*r = *r.WithContext(context.WithValue(r.Context(), ctxdata.Identify, claims[ctxdata.Identify]))
	return true
}

func (j *JwtAuth) UserId(r *http.Request) string {
	return ctxdata.GetUid(r.Context())
}

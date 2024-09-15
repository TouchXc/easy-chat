package svc

import (
	"easy-chat/apps/im/rpc/imclient"
	"easy-chat/apps/social/api/internal/config"
	"easy-chat/apps/social/api/internal/middleware"
	"easy-chat/apps/social/rpc/socialclient"
	"easy-chat/apps/user/rpc/userclient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                config.Config
	LimitMiddleware       rest.Middleware
	IdempotenceMiddleware rest.Middleware
	socialclient.Social   // 社交服务客户端
	userclient.User       // 用户服务客户端
	imclient.Im           // 即时通讯服务客户端
	*redis.Redis          // Redis 客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		LimitMiddleware:       middleware.NewLimitMiddleware().Handle,
		IdempotenceMiddleware: middleware.NewIdempotenceMiddleware().Handle,
		Social:                socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:                  userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Im:                    imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
	}
}

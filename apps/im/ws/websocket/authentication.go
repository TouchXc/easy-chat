//完成im鉴权工作

package websocket

import (
	"fmt"
	"net/http"
	"time"
)

type Authentication interface {
	Auth(w http.ResponseWriter, r *http.Request) bool
	UserId(r *http.Request) string
}

type authentication struct {
}

func (*authentication) Auth(w http.ResponseWriter, r *http.Request) bool {
	return true
}

//根据请求参数获取用户id信息

func (*authentication) UserId(r *http.Request) string {
	query := r.URL.Query()
	//这里直接去解析请求中是否有用户id的设置，如果没有则用时间戳去生成唯一的用户id标识符
	if query != nil && query["userId"] != nil {
		return fmt.Sprintf("%v", query["userId"])
	}
	//这里是用时间戳生成用户id标识符
	return fmt.Sprintf("%v", time.Now().UnixMilli())
}

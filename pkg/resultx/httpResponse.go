package resultx

import (
	"context"
	"easy-chat/pkg/xerr"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	zerr "github.com/zeromicro/x/errors"
	"google.golang.org/grpc/status"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) *Response {
	return &Response{
		Code: 200,
		Msg:  "",
		Data: data,
	}
}
func Fail(code int, err string) *Response {
	return &Response{
		Code: code,
		Msg:  err,
		Data: nil,
	}
}

func OkHandler(_ context.Context, v interface{}) any {
	return Success(v)
}

func ErrHandler(name string) func(ctx context.Context, err error) (int, any) {
	return func(ctx context.Context, err error) (int, any) {
		errCode := xerr.SERVER_COMMON_ERROE
		errMsg := xerr.ErrMsg(errCode)
		// 获取错误的根本原因
		causeErr := errors.Cause(err)
		// 处理自定义错误类型
		if e, ok := causeErr.(*zerr.CodeMsg); ok {
			errCode = e.Code
			errMsg = e.Msg
		} else {
			// 处理gRPC错误
			if gstatus, ok := status.FromError(causeErr); ok {
				errCode = int(gstatus.Code())
				errMsg = gstatus.Message()
			}
		}
		// 记录错误日志
		logx.WithContext(ctx).Errorf("【%s】 err: %v", name, err)

		// 返回HTTP状态码和错误响应
		return http.StatusBadRequest, Fail(errCode, errMsg)
	}
}

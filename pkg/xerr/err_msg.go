package xerr

var codeText = map[int]string{
	SERVER_COMMON_ERROE: "服务器异常，稍后再尝试",
	REQUEST_PARAM_ERROR: "请求参数有误",
	DB_ERROR:            "数据库繁忙，稍后再尝试",
}

func ErrMsg(errCode int) string {
	if msg, ok := codeText[errCode]; ok {
		return msg
	}
	return codeText[SERVER_COMMON_ERROE]
}

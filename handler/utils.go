package handler

import (
	"github.com/labstack/echo/v4"
)

type code string

// return code:

// common
const (
	codeOK           code = "OK"
	codeInvalidParam code = "INVALID_PARAM" // 参数错误
)

// user
const (
	codeUserExist    code = "USER_EXIST"
	codeUserNotExist code = "USER_NOT_EXIST"
	codeAuthError    code = "AUTH_ERROR"
)

func commonJSON(c echo.Context, httpCode int, code code, msg string, d interface{}) error {
	resp := commonResp{
		Code: code,
		Msg:  msg,
		Data: d,
	}

	return c.JSON(httpCode, resp)
}

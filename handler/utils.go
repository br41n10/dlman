package handler

import "github.com/labstack/echo/v4"

func commonJSON(c echo.Context, httpCode int, code string, msg string, d interface{}) error {
	resp := commonResp{
		Code: code,
		Msg:  msg,
		Data: d,
	}

	return c.JSON(httpCode, resp)
}

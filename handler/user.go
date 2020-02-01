package handler

import (
	"dlman/config"
	"dlman/data"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"time"
)

type signInResp struct {
	JwtToken string `json:"jwt_token"`
}

func SignUp(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// 检查参数输入
	if len(email) < 5 || len(password) < 8 {
		return commonJSON(c, http.StatusOK, codeInvalidParam, "邮箱或密码不符合要求", nil)
	}

	// 判断用户是否已经注册
	exist, err := data.IsUserExistByEmail(email)
	if err != nil {
		log.Errorf("SignUp|IsUserExistByEmail|%v", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	if exist {
		return commonJSON(c, http.StatusOK, codeUserExist, "用户已存在", nil)
	}

	// 新建用户
	pe, err := data.EncryptPassword(password)
	if err != nil {
		log.Errorf("SignUp|EncryptPassword|%v", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	_, err = data.CreateUserByEmail(email, pe, email)
	if err != nil {
		log.Errorf("SignUp|CreateUserByEmail|%v", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	// TODO: 这里需要直接返回jwt的一套东西，让用户处于登陆状态

	return commonJSON(c, http.StatusOK, codeOK, "注册成功，请登录", nil)
}

func SignIn(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// 检查参数输入
	if len(email) < 5 || len(password) < 8 {
		return commonJSON(c, http.StatusOK, codeInvalidParam, "邮箱或密码不符合要求", nil)
	}

	// 是否存在该用户
	exist, err := data.IsUserExistByEmail(email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
	}
	if !exist {
		return commonJSON(c, http.StatusOK, codeUserNotExist, "用户不存在", nil)
	}

	// 取出用户
	user, err := data.GetUserByEmail(email)
	if err != nil {
		log.Errorf("SignIn|GetUserByEmail|%v", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	// check password
	valid := data.CheckPassword(password, user.PasswordEncrypted.String)
	if !valid {
		return commonJSON(c, http.StatusOK, codeAuthError, "认证错误", nil)
	}

	// 通过了身份认证

	// jwt
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Id
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()

	t, err := token.SignedString([]byte(config.JwtKey))
	if err != nil {
		log.Errorf("SignIn|SignedString|%v", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	resp := signInResp{JwtToken: t}

	return commonJSON(c, http.StatusOK, codeOK, "登陆成功", resp)
}

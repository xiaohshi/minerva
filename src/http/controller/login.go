package controller

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"minerva/src/common"
	"minerva/src/logic"
	"net/http"
	"time"
)

type LoginController struct {
}

// jwt加密的数据结构体
type JwtCustomClaims struct {
	Email string `json:"email"`
	IsGod bool   `json:"admin"`
	jwt.StandardClaims
	ExpiresAt interface{}
}

var authLogic logic.AuthLogic = logic.AuthLogic{}

/**
 *  登录接口
 *	POST /minerva/login
 *
 *  @param ctx echo.Context
 *  @return error error
 */
func (login LoginController) Login(ctx echo.Context) error {

	// 用户名
	email := ctx.FormValue("email")
	// 密码
	password := ctx.FormValue("password")

	// 验证用户名密码
	error := verify(email, password)
	if error != nil {
		common.Logger.WithFields(logrus.Fields{
			"file":   "http/controller/login.go",
			"method": "Login",
			"type":   "verify error",
		}).Errorln(error.Error())
		//(message)
		return echo.NewHTTPError(http.StatusInternalServerError, error.Error())
	}

	// 生成token的必要数据
	data := map[string]interface{}{
		"email": email,
		"isGod": true,
	}

	// 生成token
	token, error := generateToken(data)

	if error != nil {
		common.Logger.WithFields(logrus.Fields{
			"file":   "http/controller/login.go",
			"method": "Login",
			"type":   "generateToken error",
		}).Errorln(error.Error())
		//	common.Logger.Errorln("login.go #Login generateToken error :", error)
		return echo.NewHTTPError(http.StatusInternalServerError, error.Error())
	}

	// 设置session
	authLogic.SetLoginInfo(ctx, email)

	return ctx.JSON(http.StatusOK, echo.Map{
		"token": token,
	})

}

/**
校验用户名和密码
*/
func verify(email string, password string) error {
	// 后续考虑第三方认证。。。。todo
	result, error := authLogic.Verify(email, password)
	if !result {
		common.Logger.WithFields(logrus.Fields{
			"file":   "http/controller/login.go",
			"method": "verify",
			"type":   "verify error",
		}).Errorln(error)
		// common.Logger.Errorln("login.go# verify error:", error)
		return error
	}

	return nil
}

/**
利用JWT生成token
   JwtCustomClaims 当参数?
*/
func generateToken(data map[string]interface{}) (string, error) {
	// jwt private_key
	var jwtKey string = common.Global.JWTConfig.Key
	// 过期时间 -- 必须转化成time.Duration格式 不然会抛异常
	var d time.Duration = common.Global.JWTConfig.Expire

	token := jwt.New(jwt.SigningMethodHS256)

	claims := &JwtCustomClaims{
		Email:     data["email"].(string),
		IsGod:     data["isGod"].(bool),
		ExpiresAt: time.Now().Add(d).Unix(),
	}

	// 设置claims
	token.Claims = claims

	//	通过claims生成token
	reallyToken, err := token.SignedString([]byte(jwtKey))

	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// jwt签名
	//reallyToken, err := token.SignedString([]byte(jwtKey))

	if err != nil {
		common.Logger.WithFields(logrus.Fields{
			"file":   "http/controller/login.go",
			"method": "generateToken",
			"type":   "generateToken error",
		}).Errorln(err)
		//	common.Logger.Errorln("login.go#generateToken error:", err)
		return "", err
	}

	return reallyToken, nil
}

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"com.github.gin-common/app/service"

	"com.github.gin-common/common/tools/jwt_tool"

	"com.github.gin-common/app/model"

	"com.github.gin-common/common/exceptions"

	"com.github.gin-common/app/exception"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

type AuthMiddleware struct {
	rdb         *redis.Client
	ctx         context.Context
	userService service.UserService
}

func (middleware *AuthMiddleware) Init(rdb *redis.Client, userService service.UserService, ctx context.Context) {
	middleware.rdb = rdb
	middleware.userService = userService
	middleware.ctx = ctx
}

func (middleware *AuthMiddleware) Before(ctx *gin.Context) (err error) {
	token := extractToken(ctx.Request)
	if token == "" {
		return exceptions.GetDefinedErrors(exception.TokenInvalid)
	}
	var claims jwt.MapClaims
	claims, err = jwt_tool.TokenClaims(token)

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return exceptions.GetDefinedErrors(exception.TokenInvalid)
			}
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return exceptions.GetDefinedErrors(exception.TokenExpired)
			}
			return exceptions.GetDefinedErrors(exception.TokenInvalid)
		}
		return exceptions.GetDefinedErrors(exception.TokenInvalid)
	}

	tokenUUID, ok := claims["tokenUUID"].(string)
	if !ok {
		return exceptions.GetDefinedErrors(exception.TokenInvalid)
	}

	var val string
	val, err = middleware.rdb.Get(middleware.ctx, fmt.Sprintf("accessToken:%s", tokenUUID)).Result()
	if err != nil {
		if err == redis.Nil {
			// token失效
			return exceptions.GetDefinedErrors(exception.TokenExpired)
		}
		return
	}
	var userData map[string]uint
	err = json.Unmarshal([]byte(val), &userData)
	if err != nil {
		return
	}
	//根据userID获取用户信息
	var user *model.User
	user, err = middleware.userService.GetUserInfoById(userData["userId"])
	if err != nil {
		return
	}
	ctx.Set("userInfo", map[string]interface{}{
		"user":      user,
		"tokenUUID": tokenUUID,
	})
	return
}

func (middleware *AuthMiddleware) After(ctx *gin.Context) (err error) {
	return
}

func (middleware *AuthMiddleware) DeniedBeforeAbortContext() bool {
	return false
}

func (middleware *AuthMiddleware) AllowAfterAbortContext() bool {
	return false
}

func GetAuthedUserInfo(ctx *gin.Context) (userInfo map[string]interface{}, err error) {
	// 获取用户信息
	value, exists := ctx.Get("userInfo")
	if !exists {
		err = errors.New("未从ctx中获取到用户信息")
		return
	}
	var ok bool
	userInfo, ok = value.(map[string]interface{})
	if !ok {
		err = errors.New("未获取到用户信息")
		return
	}
	return
}

type DeactivatedAbortMiddleware struct{}

func (middleware *DeactivatedAbortMiddleware) Before(ctx *gin.Context) (err error) {
	var userInfo map[string]interface{}

	userInfo, err = GetAuthedUserInfo(ctx)
	if err != nil {
		return
	}
	user, ok := userInfo["user"].(*model.User)
	if !ok {
		err = exceptions.GetDefinedErrors(exceptions.ServerError)
		return
	}
	if user.ActivateStatus == false {
		return exceptions.GetDefinedErrors(exception.UserDeactivated)
	}
	return
}

func (middleware *DeactivatedAbortMiddleware) After(ctx *gin.Context) (err error) {
	return
}

func (middleware *DeactivatedAbortMiddleware) DeniedBeforeAbortContext() bool {
	return false
}

func (middleware *DeactivatedAbortMiddleware) AllowAfterAbortContext() bool {
	return false
}

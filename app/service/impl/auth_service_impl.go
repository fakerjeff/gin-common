package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"com.github.gin-common/app/model"
	"com.github.gin-common/util"

	"com.github.gin-common/app/exception"
	"com.github.gin-common/app/service"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/tools/jwt_tool"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type AuthServiceImpl struct {
	ctx         context.Context
	rdb         *redis.Client
	userService service.UserService
}

func (authService *AuthServiceImpl) Init(ctx context.Context, rdb *redis.Client, userService service.UserService) {
	authService.ctx = ctx
	authService.rdb = rdb
	authService.userService = userService
}

type Option struct {
}

func (authService *AuthServiceImpl) Login(username string, password string) (token service.TokenString, err error) {
	var user *model.User
	user, err = authService.userService.GetUserInfoByUserName(username)
	if err != nil {
		err = exceptions.GetDefinedErrors(exception.UserNameOrPassInvalid)
		return
	}
	if user.ActivateStatus == false {
		err = exceptions.GetDefinedErrors(exception.UserDeactivated)
		return
	}

	if user.CheckPass(password) {
		// 生成token
		var e error
		token, e = authService.saveAuth(user.ID)
		if e != nil {
			err = exceptions.GetDefinedErrors(exception.LoginFailed)
			return
		}
	} else {
		err = exceptions.GetDefinedErrors(exception.UserNameOrPassInvalid)
		return
	}
	return
}

func (authService *AuthServiceImpl) saveAuth(userId uint) (service.TokenString, error) {
	// 创建token,并存入redis
	tokenUUID := uuid.New().String()
	claims := jwt.MapClaims{
		"tokenUUID": tokenUUID,
	}
	tokenExpireSecond, err := strconv.Atoi(util.GetDefaultEnv("ACCESS_TOKEN_EXPIRE", strconv.Itoa(2*60*60)))
	if err != nil {
		return "", err
	}
	var tokenInfo *jwt_tool.TokenInfo
	tokenInfo, err = jwt_tool.CreateToken(claims, tokenExpireSecond)
	if err != nil {
		return "", err
	}
	// use UUID4 to create uuid
	expiredAt := time.Unix(tokenInfo.ExpiredAt, 0)
	v, e := json.Marshal(map[string]uint{
		"userId": userId,
	})
	if e != nil {
		return "", e
	}

	if err := authService.rdb.Set(authService.ctx, fmt.Sprintf("accessToken:%s", tokenUUID), v, expiredAt.Sub(time.Now())).Err(); err != nil {
		return "", err
	}
	return service.TokenString(tokenInfo.Token), err
}

func (authService *AuthServiceImpl) Logout(sessionID string) error {
	_, err := authService.rdb.Get(authService.ctx, fmt.Sprintf("accessToken:%s", sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}
	_, err = authService.rdb.Del(authService.ctx, fmt.Sprintf("accessToken:%s", sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}
	return nil
}

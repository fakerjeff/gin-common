package auth

import (
	"com.github.gin-common/app/middleware"
	"com.github.gin-common/app/service"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type LogoutController struct {
	authService service.AuthService
}

func (controller *LogoutController) Init(authService service.AuthService) {
	controller.authService = authService
}

func (controller *LogoutController) logout(context *gin.Context) (data *resp.Response, err error) {
	var userInfo map[string]interface{}

	userInfo, err = middleware.GetAuthedUserInfo(context)
	if err != nil {
		return
	}
	tokenUUID, ok := userInfo["tokenUUID"].(string)
	if ok && tokenUUID != "" {
		err = controller.authService.Logout(tokenUUID)
		if err != nil {
			return
		}
	} else {
		err = exceptions.GetDefinedErrors(exceptions.ServerError)
		return
	}
	data = controllers.Success(gin.H{})
	return
}

func (controller *LogoutController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.logout(context)
}

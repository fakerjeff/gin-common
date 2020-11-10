package auth

import (
	"com.github.gin-common/app/middleware"
	"com.github.gin-common/app/model"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type CurrentUserController struct {
}

func (controller *CurrentUserController) currentUser(context *gin.Context) (data *resp.Response, err error) {
	var userInfo map[string]interface{}

	userInfo, err = middleware.GetAuthedUserInfo(context)
	if err != nil {
		return
	}
	user, ok := userInfo["user"].(*model.User)

	if ok && user != nil {
		data = controllers.Success(gin.H{
			"user": user,
		})
	} else {
		err = exceptions.GetDefinedErrors(exceptions.ServerError)
		return
	}
	return
}

func (controller *CurrentUserController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.currentUser(context)
}

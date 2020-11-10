package user

import (
	"errors"
	"strconv"

	"com.github.gin-common/app/model"

	"com.github.gin-common/app/service"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type GetUserInfoController struct {
	userService service.UserService
}

func (controller *GetUserInfoController) Init(userService service.UserService) {
	controller.userService = userService
}

func (controller *GetUserInfoController) getUserInfo(context *gin.Context) (data *resp.Response, err error) {
	// 根据ID获取用户信息
	userID := context.Param("userID")
	intUserID, e := strconv.Atoi(userID)
	if e != nil {
		err = exceptions.NewError(exceptions.BadRequest, exceptions.WithError(errors.New("传入正确的userID")))()
		return
	}
	var user *model.User
	user, err = controller.userService.GetUserInfoById(uint(intUserID))

	if err != nil {
		return
	}

	data = controllers.Success(gin.H{
		"user": user,
	})
	return
}

func (controller *GetUserInfoController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.getUserInfo(context)
}

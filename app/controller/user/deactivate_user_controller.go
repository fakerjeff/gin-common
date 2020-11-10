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

type DeActivateUserController struct {
	userService service.UserService
}

func (controller *DeActivateUserController) Init(userService service.UserService) {
	controller.userService = userService
}

func (controller *DeActivateUserController) deActivateUser(context *gin.Context) (data *resp.Response, err error) {
	var user *model.User
	userID := context.Param("userID")
	intUserID, e := strconv.Atoi(userID)
	if e != nil {
		err = exceptions.NewError(exceptions.BadRequest, exceptions.WithError(errors.New("传入正确的userID")))()
		return
	}

	user, err = controller.userService.DeactivateUser(uint(intUserID))
	if err != nil {
		return
	}
	data = controllers.Success(gin.H{
		"user": user,
	})
	return
}

func (controller *DeActivateUserController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.deActivateUser(context)
}

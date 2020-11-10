package user

import (
	"errors"
	"strconv"

	"com.github.gin-common/app/form"
	"com.github.gin-common/app/service"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type ChangePasswordController struct {
	userService    service.UserService
	changePassForm *form.ChangePassForm
}

func (controller *ChangePasswordController) Init(userService service.UserService, changePassForm *form.ChangePassForm) {
	controller.userService = userService
	controller.changePassForm = changePassForm
}

func (controller *ChangePasswordController) changePassword(context *gin.Context) (data *resp.Response, err error) {
	userID := context.Param("userID")

	intUserID, e := strconv.Atoi(userID)
	if e != nil {
		err = exceptions.NewError(exceptions.BadRequest, exceptions.WithError(errors.New("传入正确的userID")))()
		return
	}
	if e := context.ShouldBindJSON(controller.changePassForm); e != nil {
		err = e
		return
	}
	err = controller.userService.ChangePassword(uint(intUserID), controller.changePassForm.OldPassword, controller.changePassForm.NewPassword)
	if err != nil {
		return
	}

	data = controllers.Success(gin.H{})
	return
}

func (controller *ChangePasswordController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.changePassword(context)
}

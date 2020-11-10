package user

import (
	"errors"
	"strconv"

	"com.github.gin-common/common/resp"

	"com.github.gin-common/common/exceptions"

	"com.github.gin-common/common/controllers"

	"com.github.gin-common/app/form"
	"com.github.gin-common/app/model"
	"com.github.gin-common/app/service"
	"github.com/gin-gonic/gin"
)

type EditUserController struct {
	updateUserForm *form.UpdateUserForm
	userService    service.UserService
}

func (controller *EditUserController) Init(updateUserForm *form.UpdateUserForm, userService service.UserService) {
	controller.updateUserForm = updateUserForm
	controller.userService = userService
}

func (controller *EditUserController) editUser(context *gin.Context) (data *resp.Response, err error) {
	userID := context.Param("userID")
	intUserID, e := strconv.Atoi(userID)
	if e != nil {
		err = exceptions.NewError(exceptions.BadRequest, exceptions.WithError(errors.New("传入正确的userID")))()
		return
	}
	if e := context.ShouldBindJSON(controller.updateUserForm); e != nil {
		err = e
		return
	}
	var user *model.User
	user, err = controller.userService.UpdateUser(uint(intUserID), model.User{
		Name:     controller.updateUserForm.Name,
		Email:    controller.updateUserForm.Email,
		Username: controller.updateUserForm.UserName,
	})
	if err != nil {
		return
	}
	data = controllers.Success(gin.H{
		"user": user,
	})
	return
}

func (controller *EditUserController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.editUser(context)
}

package user

import (
	"com.github.gin-common/common/resp"

	"com.github.gin-common/common/controllers"

	"com.github.gin-common/app/form"
	"com.github.gin-common/app/model"
	"com.github.gin-common/app/service"
	"github.com/gin-gonic/gin"
)

type CreateUserController struct {
	createUserForm *form.CreateUserForm
	userService    service.UserService
}

func (controller *CreateUserController) Init(createUserForm *form.CreateUserForm, userService service.UserService) {
	controller.createUserForm = createUserForm
	controller.userService = userService
}

func (controller *CreateUserController) createUser(context *gin.Context) (data *resp.Response, err error) {
	if e := context.ShouldBindJSON(controller.createUserForm); e != nil {
		err = e
		return
	}
	var user *model.User
	user, err = controller.userService.CreateUser(&model.User{
		Name:     controller.createUserForm.Name,
		Email:    controller.createUserForm.Email,
		Username: controller.createUserForm.UserName,
	}, controller.createUserForm.Password)

	if err != nil {
		return
	}
	data = controllers.Success(gin.H{
		"user": user,
	})
	return
}

func (controller *CreateUserController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.createUser(context)
}

package auth

import (
	"com.github.gin-common/app/form"
	"com.github.gin-common/app/service"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type LoginController struct {
	loginForm   *form.LoginForm
	authService service.AuthService
}

func (controller *LoginController) Init(loginForm *form.LoginForm, authService service.AuthService) {
	controller.loginForm = loginForm
	controller.authService = authService
}

func (controller *LoginController) login(context *gin.Context) (data *resp.Response, err error) {
	if e := context.ShouldBindJSON(controller.loginForm); e != nil {
		return nil, e
	}
	var token service.TokenString
	token, err = controller.authService.Login(controller.loginForm.UserName, controller.loginForm.Password)
	return controllers.Success(gin.H{
		"token": token,
	}), err
}

func (controller *LoginController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.login(context)
}

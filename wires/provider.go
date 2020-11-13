package wires

import (
	"context"
	"strconv"
	"time"

	"com.github.gin-common/common/controllers"

	"com.github.gin-common/tools/redis_tool"

	"github.com/go-redis/redis/v8"

	"com.github.gin-common/app/middleware"

	authController "com.github.gin-common/app/controller/auth"
	userController "com.github.gin-common/app/controller/user"

	"com.github.gin-common/app/service/impl"

	"com.github.gin-common/app/service"

	"com.github.gin-common/app/form"

	"com.github.gin-common/tools/db_tool"

	"com.github.gin-common/util"
	"gorm.io/gorm"
)

type gormSessionTimeout time.Duration
type timeoutGormContext context.Context

func provideTimeoutGormSession(context timeoutGormContext) *gorm.DB {
	session := db_tool.GetDB().Session(&gorm.Session{Context: context, WithConditions: false})
	return session
}

func provideTimeoutGormContext(timeout gormSessionTimeout) timeoutGormContext {
	timeoutContext, _ := context.WithTimeout(context.Background(), time.Duration(timeout))
	return timeoutContext
}

func provideGormSessionTimeout() gormSessionTimeout {
	strTimeout := util.GetDefaultEnv("GORM_CONTEXT_TIMEOUT", "5")
	timeout, err := strconv.Atoi(strTimeout)
	if err != nil {
		return gormSessionTimeout(5 * time.Second)
	}
	return gormSessionTimeout(time.Duration(timeout) * time.Second)
}

func provideCreateUserController(createUserForm *form.CreateUserForm, userService service.UserService) controllers.Controller {
	c := &userController.CreateUserController{}
	c.Init(createUserForm, userService)
	return c
}

func provideCreateUserForm() *form.CreateUserForm {
	return &form.CreateUserForm{}
}

func provideUserService(session *gorm.DB) *impl.UserServiceImpl {
	serviceImpl := &impl.UserServiceImpl{}
	serviceImpl.Init(session)
	return serviceImpl
}

func provideUpdateUserForm() *form.UpdateUserForm {
	return &form.UpdateUserForm{}
}

func provideEditUserController(updateUserForm *form.UpdateUserForm, userService service.UserService) controllers.Controller {
	controller := &userController.EditUserController{}
	controller.Init(updateUserForm, userService)
	return controller
}

func provideDeleteUserController(userService service.UserService) controllers.Controller {
	controller := &userController.DeleteUserController{}
	controller.Init(userService)
	return controller
}

func provideActivateUserController(userService service.UserService) controllers.Controller {
	controller := &userController.ActivateUserController{}
	controller.Init(userService)
	return controller
}

func provideDeActivateUserController(userService service.UserService) controllers.Controller {
	controller := &userController.DeActivateUserController{}
	controller.Init(userService)
	return controller
}

func provideGetUserInfoController(userService service.UserService) controllers.Controller {
	controller := &userController.GetUserInfoController{}
	controller.Init(userService)
	return controller
}

func provideChangePasswordController(changePassForm *form.ChangePassForm, userService service.UserService) controllers.Controller {
	controller := &userController.ChangePasswordController{}
	controller.Init(userService, changePassForm)
	return controller
}

func provideChangePassForm() *form.ChangePassForm {
	return &form.ChangePassForm{}
}

func provideRedisContext() context.Context {
	return context.Background()
}

func provideRedisRdb() *redis.Client {
	return redis_tool.GetGinServerRdb()
}

func provideAuthMiddleware(ctx context.Context, rdb *redis.Client, userService service.UserService) controllers.MiddleWare {
	mid := &middleware.AuthMiddleware{}
	mid.Init(rdb, userService, ctx)
	return mid
}

func provideDeactivatedAbortMiddleware() controllers.MiddleWare {
	return &middleware.DeactivatedAbortMiddleware{}
}

func provideLoginForm() *form.LoginForm {
	return &form.LoginForm{}
}

func provideAuthService(ctx context.Context, rdb *redis.Client, userService service.UserService) *impl.AuthServiceImpl {
	serviceImpl := &impl.AuthServiceImpl{}
	serviceImpl.Init(ctx, rdb, userService)
	return serviceImpl
}

func provideLoginController(loginForm *form.LoginForm, authService service.AuthService) controllers.Controller {
	controller := &authController.LoginController{}
	controller.Init(loginForm, authService)
	return controller
}

func provideLogoutController(authService service.AuthService) controllers.Controller {
	controller := &authController.LogoutController{}
	controller.Init(authService)
	return controller
}

func provideCurrentUserController() controllers.Controller {
	return &authController.CurrentUserController{}
}

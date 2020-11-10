package user

import (
	"errors"
	"strconv"

	"com.github.gin-common/app/service"
	"com.github.gin-common/common/controllers"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
)

type DeleteUserController struct {
	userService service.UserService
}

func (controller *DeleteUserController) Init(userService service.UserService) {
	controller.userService = userService
}

func (controller *DeleteUserController) deleteUser(context *gin.Context) (data *resp.Response, err error) {
	// 删除用户
	userID := context.Param("userID")
	intUserID, e := strconv.Atoi(userID)
	if e != nil {
		err = exceptions.NewError(exceptions.BadRequest, exceptions.WithError(errors.New("传入正确的userID")))()
		return
	}
	err = controller.userService.DeleteUser(uint(intUserID))
	if err != nil {
		return
	}
	data = controllers.Success(make(map[string]interface{}))
	return
}

func (controller *DeleteUserController) DoRequest(context *gin.Context) (data *resp.Response, err error) {
	return controller.deleteUser(context)
}

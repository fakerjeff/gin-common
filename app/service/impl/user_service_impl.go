package impl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"com.github.gin-common/util"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"

	"com.github.gin-common/common/caches"

	"com.github.gin-common/common/models"

	"com.github.gin-common/app/exception"
	"com.github.gin-common/app/model"
	"com.github.gin-common/common/exceptions"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserServiceImpl struct {
	session *gorm.DB
	rdb     *redis.Client
	ctx     context.Context
	logger  zap.Logger
}

func (service *UserServiceImpl) Init(session *gorm.DB, rdb *redis.Client, ctx context.Context, logger zap.Logger) {
	service.session = session
	service.rdb = rdb
	service.ctx = ctx
	service.logger = logger
}

func (service *UserServiceImpl) CreateUser(user *model.User, password string) (*model.User, error) {
	// 创建用户
	e := user.SetPass(password)
	if e != nil {
		return nil, exceptions.GetDefinedErrors(exception.UserCreateFailed)
	}
	user.ActivateAt = time.Now()
	result := service.session.Create(user)
	if result.Error != nil {
		if err, ok := result.Error.(*mysql.MySQLError); ok {
			if err.Number == uint16(1062) {
				return nil, exceptions.GetDefinedErrors(exception.UserNameDuplicate)
			}
		}
		return nil, exceptions.GetDefinedErrors(exception.UserCreateFailed)
	}

	return user, nil
}

func (service *UserServiceImpl) UpdateUser(id uint, updateInfo model.User) (*model.User, error) {
	// 更新用户

	var user, err = service.GetUserInfoById(id)
	if err != nil {
		return nil, err
	}
	if result := service.session.Model(user).Updates(updateInfo); result.Error != nil {
		return nil, exceptions.GetDefinedErrors(exception.UserUpdateFailed)
	}
	return user, nil
}

func (service *UserServiceImpl) DeleteUser(id uint) error {
	// 删除用户
	user, err := service.GetUserInfoById(id)
	if err != nil {
		return err
	}
	if result := models.Delete(service.session, id, user); result.Error != nil {
		return exceptions.GetDefinedErrors(exception.DeleteUserFailed)
	}
	return nil
}

func (service *UserServiceImpl) ActivateUser(id uint) (*model.User, error) {
	// 启用用户
	user, err := service.GetUserInfoById(id)
	if err != nil {
		return nil, err
	}
	if result := models.Activate(service.session, id, user); result.Error != nil {
		return nil, exceptions.GetDefinedErrors(exception.ActivateUserFailed)
	}
	return user, nil
}

func (service *UserServiceImpl) DeactivateUser(id uint) (*model.User, error) {
	// 禁用用户
	user, err := service.GetUserInfoById(id)
	if err != nil {
		return nil, err
	}
	if result := models.Deactivate(service.session, id, user); result.Error != nil {
		return nil, exceptions.GetDefinedErrors(exception.DeActivateUserFailed)
	}
	return user, nil
}

func (service *UserServiceImpl) GetUserInfoById(id uint) (*model.User, error) {
	user := &model.User{}

	bridge := func(process func(id uint) (*model.User, error), id uint) func() (interface{}, error) {
		return func() (interface{}, error) {
			user, err := process(id)
			return user, err
		}
	}
	redisCache := new(caches.RedisCache)
	redisCache.Init(service.rdb, service.ctx)

	redisCacheProvideOption := caches.RedisCacheProvideOption{
		RedisCache: redisCache,
	}
	expiresOption := caches.CacheExpiresOption(5 * time.Minute)
	tool := new(util.SerializeTool)
	tool.Init(service.logger)

	serializerOption := caches.SerializerOption{
		Serializer: tool,
	}
	var err error
	var result interface{}
	result, err = caches.CacheEnable(bridge(service.getUserInfoById, id), user, caches.CacheKeyOption(fmt.Sprintf("user:%d", id)),
		redisCacheProvideOption, expiresOption, serializerOption)
	if err != nil {
		return nil, err
	}
	user = result.(*model.User)
	return user, nil
}

func (service *UserServiceImpl) getUserInfoById(id uint) (*model.User, error) {
	user := &model.User{}
	if result := service.session.Where("ID=?", id).Take(user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exceptions.GetDefinedErrors(exception.UserNotFound)
		}
		return nil, result.Error
	}
	return user, nil
}

func (service *UserServiceImpl) GetUserInfoByUserName(username string) (*model.User, error) {
	// 通过用户名获取用户
	user := &model.User{}
	if result := service.session.Where("username=?", username).Take(user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exceptions.GetDefinedErrors(exception.UserNotFound)
		}
		return nil, result.Error
	}
	return user, nil
}

func (service *UserServiceImpl) ChangePassword(id uint, oldPass string, newPass string) error {
	// 修改密码
	user, err := service.GetUserInfoById(id)
	if err != nil {
		return err
	}
	if user.CheckPass(oldPass) {
		if err := user.SetPass(newPass); err != nil {
			return exceptions.GetDefinedErrors(exception.ChangePassFailed)
		}
	} else {
		return exceptions.GetDefinedErrors(exception.OldPassInvalid)
	}
	service.session.Save(user)
	return nil
}

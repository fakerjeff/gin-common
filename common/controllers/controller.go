package controllers

import (
	"errors"

	"com.github.gin-common/common/consts"
	"com.github.gin-common/common/exceptions"
	"com.github.gin-common/common/loggers/gin_logger"
	"com.github.gin-common/common/resp"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ControllerFunc func() Controller
type MiddlewareFunc func() MiddleWare

type MiddleWare interface {
	Before(context *gin.Context) (err error)
	After(context *gin.Context) (err error)
	DeniedBeforeAbortContext() bool
	AllowAfterAbortContext() bool
}

type Controller interface {
	DoRequest(context *gin.Context) (data *resp.Response, err error)
}

func getErrRender(err error) resp.RenderFunc {
	// 根据异常类型获取render
	var r resp.RenderFunc
	if e, ok := err.(*exceptions.ApiError); ok {
		r = e.RenderJSONFunc()
	} else if e, ok := err.(validator.ValidationErrors); ok {
		r = exceptions.NewError(exceptions.BadRequest, exceptions.WithValidateError(e))().RenderJSONFunc()
	} else {
		if gin.IsDebugging() {
			r = exceptions.NewError(exceptions.ServerError, exceptions.WithError(err))().RenderJSONFunc()
		} else {
			r = exceptions.NewError(exceptions.ServerError)().RenderJSONFunc()
		}
		gin_logger.Log.Error(err.Error())
	}

	return r
}

func ControllerHandler(controllerFunc ControllerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		controller := controllerFunc()
		var r resp.RenderFunc
		data, err := controller.DoRequest(context)
		if err != nil {
			r = getErrRender(err)
		} else {
			r = data.RenderJSONFunc()
		}
		r(context)
	}
}

const (
	phaseBefore = 1
	phaseAfter  = 2
)

func processMiddlewareFunc(phase int, middleware MiddleWare, context *gin.Context) (goOn bool) {
	var allow bool
	var processFunc func(context *gin.Context) (err error)
	switch phase {
	case phaseBefore:
		processFunc = middleware.Before
		allow = !middleware.DeniedBeforeAbortContext()
	case phaseAfter:
		processFunc = middleware.After
		allow = middleware.AllowAfterAbortContext()
	default:
		panic(errors.New("invalid phase"))
	}
	if processFunc != nil {
		err := processFunc(context)
		if err != nil && allow {
			r := getErrRender(err)
			r(context)
			context.Abort()
			return
		}
	}
	goOn = true
	return
}

func MiddlewareHandler(middlewareFunc MiddlewareFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		mid := middlewareFunc()
		if ok := processMiddlewareFunc(phaseBefore, mid, context); !ok {
			return
		}
		context.Next()

		if ok := processMiddlewareFunc(phaseAfter, mid, context); !ok {
			return
		}
	}
}

func MakeResponse(code string, data gin.H, msg string) *resp.Response {
	return &resp.Response{
		Code: code,
		Data: data,
		Msg:  msg,
	}
}

func Success(data gin.H) *resp.Response {
	return MakeResponse(consts.HttpSuccessCode, data, consts.HttpSuccess)
}

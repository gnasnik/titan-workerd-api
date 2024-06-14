package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	err "github.com/gnasnik/titan-workerd-api/core/errors"
	"github.com/pkg/errors"
)

type JsonObject map[string]interface{}

func respJSON(v interface{}) gin.H {
	return gin.H{
		"success": true,
		"data":    v,
		"code":    0,
	}
}

func respError(e error) gin.H {
	var apiError err.ApiError
	if !errors.As(e, &apiError) {
		apiError = err.ErrUnknown
	}

	return gin.H{
		"success": false,
		"code":    apiError.Code(),
		"message": apiError.Error(),
	}
}

func respErrorWrapMessage(e error, msg string) gin.H {
	var apiError err.ApiError
	if !errors.As(e, &apiError) {
		apiError = err.ErrUnknown
	}

	return gin.H{
		"success": false,
		"code":    apiError.Code(),
		"message": fmt.Sprintf("error: %s", msg),
	}
}

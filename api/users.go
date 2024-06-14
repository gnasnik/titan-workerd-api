package api

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-workerd-api/core/dao"
	"github.com/gnasnik/titan-workerd-api/core/errors"
	"net/http"
)

func GetUserInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
		return
	}

	c.JSON(http.StatusOK, respJSON(user))
}

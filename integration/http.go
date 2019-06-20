package integration

import (
	"encoding/base64"
	"geitaidenwaMonitor/controler"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func NewHTTP(controller controler.ServiceController, access controler.Access) *gin.Engine {
	router := gin.Default()
	authOnly := router.Group("/monitor").Use(func(gctx *gin.Context) {
		const realm = "Authorization Required"
		hRealm := "Basic realm=" + strconv.Quote(Realm)
		authBase := gctx.Request.Header.Get("Authorization")
		authScheme := strings.Split(authBase, " ")
		if authScheme[0] != "Basic" || len(authScheme) != 2 {
			gctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		auth, err := base64.StdEncoding.DecodeString(authScheme[1])
		if err != nil {
			gctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		up := strings.SplitN(string(auth), ":", 2)
		if len(up) == 2 && access.Login(up[0], up[1]) == nil {
			gctx.Next()
			return
		}
		gctx.Header(WWWAuthHeader, hRealm)
		gctx.AbortWithStatus(http.StatusUnauthorized)
	})
	authOnly.GET("/run/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Run(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/stop/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Stop(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/restart/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Restart(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/status", func(gctx *gin.Context) {
		response := controller.RefreshStatus()
		gctx.IndentedJSON(http.StatusOK, response)
	})
	authOnly.GET("/status/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		status := controller.Status(name)
		gctx.IndentedJSON(http.StatusOK, status)
	})
	authOnly.GET("/log/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if log, err := controller.Log(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		} else {
			gctx.String(http.StatusOK, log)
		}
	})
	return router
}

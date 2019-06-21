package integration

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"sukauto/controler"
)

func NewHTTP(controller controler.ServiceController, access controler.Access, cors bool) *gin.Engine {
	router := gin.Default()

	if cors {
		router.Use(CORSMiddleware())
	}
	router.GET("/", func(gctx *gin.Context) {
		gctx.Redirect(http.StatusTemporaryRedirect, "public")
	})
	router.StaticFS("/public/", assetFS())

	authOnly := router.Group("/monitor").Use(func(gctx *gin.Context) {
		hRealm := "Basic realm=" + strconv.Quote(Realm)
		authBase := gctx.Request.Header.Get("Authorization")
		authScheme := strings.Split(authBase, " ")
		if authScheme[0] != "Basic" || len(authScheme) != 2 {
			gctx.Header(WWWAuthHeader, hRealm)
			gctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		auth, err := base64.StdEncoding.DecodeString(authScheme[1])
		if err != nil {
			gctx.Header(WWWAuthHeader, hRealm)
			gctx.AbortWithStatus(http.StatusUnauthorized)
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
	authOnly.GET("/update/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Update(name); err != nil {
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
	authOnly.GET("/enable/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Enable(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/disable/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Disable(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.POST("/create", func(gctx *gin.Context) {
		var newService controler.NewService
		err := gctx.BindJSON(&newService)
		if err != nil {
			gctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := controller.Create(newService); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	return router
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

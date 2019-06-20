package main

import (
	"encoding/base64"
	"geitaidenwaMonitor/controler"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var config struct {
	Bind       string `long:"bind" env:"BIND" description:"Binding address" default:":8080"`
	ConfigFile string `long:"config-file" env:"CONFIG_FILE" description:"Path to configuration file" default:"config.json"`
}

func main() {
	// TODO Call every 10 sec stat for all connected services,
	//  cli for add services to monitor
	//  cli for stop/start
	//  cli for service creation

	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}

	monitor := controler.NewServiceControllerByPath(config.ConfigFile)

	var access controler.Access = monitor

	router := gin.Default()
	authOnly := router.Group("/monitor").Use(func(gctx *gin.Context) {
		const realm = "Authorization Required"
		hRealm := "Basic realm=" + strconv.Quote(realm)
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
		gctx.Header("WWW-Authenticate", hRealm)
		gctx.AbortWithStatus(http.StatusUnauthorized)
	})
	authOnly.GET("/run/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Run(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/stop/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Stop(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/restart/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Restart(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	authOnly.GET("/status", func(gctx *gin.Context) {
		response := monitor.RefreshStatus()
		gctx.IndentedJSON(http.StatusOK, response)
	})
	authOnly.GET("/status/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		status := monitor.Status(name)
		gctx.IndentedJSON(http.StatusOK, status)
	})
	panic(router.Run(config.Bind))
}

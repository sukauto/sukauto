package main

import (
	"encoding/json"
	"geitaidenwaMonitor/controler"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

type Conf struct {
	Port string `json:"port"`
}

func main() {
	// TODO Call every 10 sec stat for all connected services,
	//  cli for add services to monitor
	//  cli for stop/start
	//  cli for service creation

	jFile, err := ioutil.ReadFile(controler.CFG_PATH)
	if err != nil {
		panic(err)
	}
	var cfg Conf
	err = json.Unmarshal(jFile, &cfg)
	if err != nil {
		panic(err)
	}

	monitor := controler.NewServiceController()

	router := gin.Default()
	router.GET("/monitor/run/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Run(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	router.GET("/monitor/stop/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Stop(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	router.GET("/monitor/restart/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := monitor.Restart(name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	router.GET("/monitor/status", func(gctx *gin.Context) {
		response := monitor.RefreshStatus()
		gctx.IndentedJSON(http.StatusOK, response)
	})
	router.GET("/monitor/status/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		status := monitor.Status(name)
		gctx.IndentedJSON(http.StatusOK, status)
	})
	panic(router.Run(cfg.Port))
}

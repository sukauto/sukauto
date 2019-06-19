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
	router.GET("/monitor/:cmd/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		cmd := strings.ToLower(strings.TrimSpace(gctx.Param("cmd")))

		if cmd == "" {
			gctx.String(http.StatusOK, "")
			return
		}
		response := cmd

		if name == "" {
			if cmd == controler.STAT {
				response = strings.Join(monitor.RefreshStatus(), " ")
			}
			gctx.String(http.StatusOK, response)
			return
		}

		if cmd == controler.STOP {
			err = monitor.Stop(name)
			if err != nil {
				panic(err)
			}
		} else if cmd == controler.RUN {
			err = monitor.Run(name)
			if err != nil {
				panic(err)
			}
		} else if cmd == controler.RESTART {
			err = monitor.Restart(name)
			if err != nil {
				panic(err)
			}
		} else if cmd == controler.STAT {
			response = monitor.Status(name)
			if err != nil {
				panic(err)
			}
		}
		gctx.String(http.StatusOK, response)
		return

	})
	panic(router.Run(cfg.Port))
}

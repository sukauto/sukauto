package integration

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sukauto/controler"
)

type CorsConfig struct {
	Allow  bool   `long:"allow" env:"ALLOW" description:"Allow CORS"`
	Origin string `long:"origin" env:"ORIGIN" description:"CORS origin host" default:"*"`
}

func NewHTTP(controller controler.ServiceController, access controler.Access, cors CorsConfig, events <-chan controler.SystemEvent) *gin.Engine {
	router := gin.Default()

	subscribe := make(chan *websocket.Conn)
	unsubscribe := make(chan *websocket.Conn)
	go func() {
		var subscribers []*websocket.Conn
		for {
			select {
			case conn := <-subscribe:
				subscribers = append(subscribers, conn)
			case conn := <-unsubscribe:
				for i, s := range subscribers {
					if s == conn {
						n := len(subscribers)
						subscribers[i] = subscribers[n-1]
						subscribers = subscribers[:n-1]
						break
					}
				}
			case event := <-events:
				payload, _ := json.MarshalIndent(event, "", "  ")
				for _, s := range subscribers {
					s.Write(payload)
				}
			}
		}
	}()

	if cors.Allow {
		router.Use(CORSMiddleware(cors))
	}
	router.GET("/", func(gctx *gin.Context) {
		gctx.Redirect(http.StatusTemporaryRedirect, "public")
	})
	router.StaticFS("/public/", assetFS())

	authOnly := router.Group("/monitor")
	authOnly.Use(func(gctx *gin.Context) {
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

	authOnly.GET("/", func(gctx *gin.Context) {
		gctx.IndentedJSON(http.StatusOK, controller.Snapshot())
	})
	authOnly.GET("/ws", gin.WrapH(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		subscribe <- ws
		io.Copy(ioutil.Discard, ws)
		unsubscribe <- ws
	})))
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
	authOnly.Use(gzip.Gzip(gzip.BestCompression)).GET("/log/:name", func(gctx *gin.Context) {
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
	authOnly.GET("/forget/:name", func(gctx *gin.Context) {
		name := strings.ToLower(strings.TrimSpace(gctx.Param("name")))
		if err := controller.Forget(name); err != nil {
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
	authOnly.POST("/attach", func(gctx *gin.Context) {
		var newService controler.PreparedService
		err := gctx.BindJSON(&newService)
		if err != nil {
			gctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := controller.Attach(newService.Name); err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	// --------------- groups section
	groups := authOnly.Group("/group")
	// all groups
	groups.GET("/", func(gctx *gin.Context) {
		gctx.IndentedJSON(http.StatusOK, controller.Groups())
	})
	// create group
	groups.POST("/:name", func(gctx *gin.Context) {
		group := gctx.Param("name")
		err := controller.Group(group)
		if err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	// remove group
	groups.DELETE("/:name", func(gctx *gin.Context) {
		group := gctx.Param("name")
		err := controller.Ungroup(group)
		if err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	// members of group
	groups.GET("/:name", func(gctx *gin.Context) {
		group := gctx.Param("name")
		gctx.IndentedJSON(http.StatusOK, controller.Members(group))
	})
	// join service to group
	groups.POST("/:name/:service", func(gctx *gin.Context) {
		group := gctx.Param("name")
		service := gctx.Param("service")
		err := controller.Join(group, service)
		if err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	// leave service from group
	groups.DELETE("/:name/:service", func(gctx *gin.Context) {
		group := gctx.Param("name")
		service := gctx.Param("service")
		err := controller.Leave(group, service)
		if err != nil {
			gctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		gctx.AbortWithStatus(http.StatusNoContent)
	})
	return router
}

func CORSMiddleware(cors CorsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cors.Origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

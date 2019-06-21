package main

import (
	"github.com/jessevdk/go-flags"
	"os"
	"sukauto/controler"
	"sukauto/integration"
)

var config struct {
	Bind       string `long:"bind" env:"BIND" description:"Binding address" default:":8080"`
	ConfigFile string `long:"config-file" env:"CONFIG_FILE" description:"Path to configuration file" default:"config.json"`
	UpdCmd     string `long:"updcmd" env:"UPDCMD" description:"command for update" default:"git pull origin master"`
	CORS       bool   `long:"cors" env:"CORS" description:"Enable CORS"`
}

func main() {
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}

	monitor := controler.NewServiceControllerByPath(config.ConfigFile, config.UpdCmd)

	var access controler.Access = monitor
	router := integration.NewHTTP(monitor, access, config.CORS)

	panic(router.Run(config.Bind))
}

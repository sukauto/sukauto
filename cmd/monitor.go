package main

import (
	"geitaidenwaMonitor/controler"
	"geitaidenwaMonitor/integration"
	"github.com/jessevdk/go-flags"
	"os"
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
	router := integration.NewHTTP(monitor, access)

	panic(router.Run(config.Bind))
}

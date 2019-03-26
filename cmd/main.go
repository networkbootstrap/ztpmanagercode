package main

// Protected by BSD 3 clause license
// This file implements the core logic for EZJunosZTP

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/networkbootstrap/ztpmanagercode/cache"
	"github.com/networkbootstrap/ztpmanagercode/cfg"
	"github.com/networkbootstrap/ztpmanagercode/rest"
)

const version = "0.0.1"

func main() {
	var configfile = flag.String("config", "./config.toml", "Configuration filename")
	var versioncheck = flag.Bool("version", false, "Version check")
	flag.Parse()

	// Deal with version check first.
	if *versioncheck == true {
		fmt.Println(version)
		os.Exit(0)
	}

	// Open file and try to parse configuration

	// Get to here, we have the configuration data ready to parse
	config := cfg.NewCfg()

	// Load instance with data
	err := config.Parse(*configfile)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	// Create cache (launches a GR) and returns communications channels
	cachesend, cachefinish := cache.Create(config.Hosts, &wg)
	wg.Add(1)
	// Create APIResponder (config service) (launches a GR) and returns communications channels
	configsend, configfinish := config.APIResponder(cachesend, *configfile, &wg)
	// Create configuration REST JSON service (launches a GR) and returns an Echo instance and error
	cfgapi, err := rest.StartCfgAPI(cachesend, configsend, config.Core.ServerPort, config.Core.HTTPUser, config.Core.HTTPPasswd)
	if err != nil {
		// Close everything else down
		close(cachefinish)
		close(configfinish)
		wg.Wait()
		os.Exit(1)
	}
	fmt.Printf("JSON API started at: %s:%v\n", config.Core.ServerURL, config.Core.ServerPort)

	// Create configuration REST static file service on port 80 (launches a GR) and returns an Echo instance and error
	fileapi, err := rest.StartStaticAPI(config.Core.FileConfigsLocation, config.Core.FileImagesLocation, config.Core.HTTPConfigsLocation, config.Core.HTTPImagesLocation)
	if err != nil {
		// Close everything else down
		cfgapi.Close()
		close(cachefinish)
		close(configfinish)
		wg.Wait()
		os.Exit(1)
	}
	fmt.Printf("File API started at: %s:80\n", config.Core.ServerURL)

	// Simple blocking
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	// Close everything else down
	fileapi.Close()
	cfgapi.Close()
	close(cachefinish)
	close(configfinish)
	wg.Wait()
}

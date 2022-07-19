package main

import (
	"flag"
	"fmt"
	"github.com/hotid/streamsurfer/internal/pkg/analyzer"
	"github.com/hotid/streamsurfer/internal/pkg/config"
	"github.com/hotid/streamsurfer/internal/pkg/helpers"
	"github.com/hotid/streamsurfer/internal/pkg/http_api"
	"github.com/hotid/streamsurfer/internal/pkg/logging"
	"github.com/hotid/streamsurfer/internal/pkg/monitor"
	"github.com/hotid/streamsurfer/internal/pkg/stats"
	"github.com/hotid/streamsurfer/internal/pkg/storage"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	var configFile string
	if runtime.GOOS == "windows" {
		configFile = "config/streamsurfer.yaml"
	} else {
		configFile = "/etc/streamsurfer.yaml"
	}

	var verbose = flag.Bool("verbose", true, "verbose output of logs")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Stream Surfer trace dumped:", r)
			if err := ioutil.WriteFile(helpers.FullPath("~/streamsurfer.trace"), r.([]byte), 0644); err != nil {
				fmt.Println("Can't write trace file!")
			}
		}
	}()

	flag.Parse()

	anotherConfig := config.InitAnotherConfig(configFile)

	storage.InitStorage()
	go logging.LogKeeper(*verbose, anotherConfig) // collect program logs and write them to file
	go stats.StatKeeper(anotherConfig)            // collect probe statistics for report builders
	go monitor.StreamMonitor(anotherConfig)       // probe logic
	go http_api.HttpAPI(anotherConfig)            // control API
	go analyzer.ProblemAnalyzer(anotherConfig)    // analyze problems related to groups of channels
	//go ProblemReporter()                          // report problems to email

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	fmt.Println("...probe service interrupted.")
}

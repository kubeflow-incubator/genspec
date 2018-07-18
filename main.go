package main

import (
	"flag"
	"os"

	"github.com/kubeflow-incubator/genspec/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	// This is needed to make `glog` believe that the flags have been parsed, otherwise
	// every log messages is prefixed by an error message.
	flag.CommandLine.Parse([]string{})

	cmd.Execute()
}

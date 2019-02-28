package main

import (
	"filekeep/config"
	"filekeep/web"
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	dumpConfig = flag.Bool("dump-config", false, "dump the default config to disk")
	configFlag = flag.String("config", "", "path to the config file")
)

var c = config.Get()

func init() {
	flag.Parse()

	if *dumpConfig {
		if err := config.Dump(); err != nil {
			logrus.WithError(err).Error("couldn't dump default config")
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *configFlag != "" {
		var err error
		c, err = config.Load(*configFlag)
		if err != nil {
			logrus.WithError(err).Error("couldn't load config")
			os.Exit(1)
		}
	}

	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debugging active")
	}
}

func main() {
	server := web.NewServer()
	logrus.WithField("address", c.Web.String()).Info("starting server")
	if err := server.ListenAndServe(); err != nil {
		logrus.WithError(err).Error("error while serving the web server")
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vlad-s/filekeep/config"
	"github.com/vlad-s/filekeep/fs"
	"github.com/vlad-s/filekeep/web"
)

var (
	loadConfig = flag.Bool("load-config", false, "Load the config from disk.")
	dumpConfig = flag.Bool("dump-config", false, "Dump the default config to disk.")
)

var l *logrus.Logger
var c *config.Config

func init() {
	l = logrus.New()
	flag.Parse()

	if *dumpConfig {
		if err := config.Dump(); err != nil {
			l.Errorln(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *loadConfig {
		if err := config.Load(); err != nil {
			l.Errorln(err)
			os.Exit(1)
		}
	}

	c = config.Get()

	if c.Debug {
		fs.Debug(c.Debug)
		fs.Log.Debugln("Debugging active")
	}
}

func main() {
	h := web.NewServer()
	fs.Log.Infof("Starting server on %s:%d", c.Listen.Addr, c.Listen.Port)
	if err := h.ListenAndServe(); err != nil {
		l.Errorln("Error during serving the web server:", err)
	}
}

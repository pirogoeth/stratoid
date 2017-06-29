package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/holster/errors"
	"gopkg.in/urfave/cli.v1"

	"github.com/pirogoeth/stratoid/config"
	"github.com/pirogoeth/stratoid/server"
)

var (
	Version   string
	BuildHash string

	commands []cli.Command = []cli.Command{}
)

func main() {
	app := cli.NewApp()
	app.Name = "stratoid"
	app.Usage = "Server component of the Stratoid mining proxy"
	app.Version = fmt.Sprintf("%s (%s)", Version, BuildHash)
	app.HideHelp = true
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, debug",
			Usage: "Be chattier about things",
		},
		cli.StringFlag{
			Name:  "listen-address, L",
			Usage: "Server listen address",
			Value: "0.0.0.0",
		},
		cli.IntFlag{
			Name:  "listen-port, P",
			Usage: "Server listen port",
			Value: 65432,
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Path to configuration file",
			Value: "./config.toml",
		},
		cli.BoolFlag{
			Name:  "profile, x",
			Usage: "Enable golang's pprof profiling extension",
		},
		cli.StringFlag{
			Name:  "profiler-address, X",
			Usage: "Address pprof HTTP server should listen on",
			Value: "localhost:6060",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		verbose := ctx.Bool("verbose")
		if verbose {
			log.SetFormatter(&log.TextFormatter{})
			log.SetOutput(os.Stderr)
			log.SetLevel(log.DebugLevel)
			log.Debug("Verbose logging enabled")
		}

		configPath, err := filepath.Abs(ctx.String("config"))
		if err != nil {
			log.WithError(err).Fatalf("Can not expand file path '%s'", ctx.String("config"))
		}

		config := &config.C{}
		if err = config.ReadConfig(configPath); err != nil {
			log.WithError(err).Fatalf("Could not read configuration file %s", configPath)
		}

		app.Metadata["config"] = config

		// ---
		launchProf := ctx.Bool("profile")
		if launchProf {
			addr := ctx.String("profiler-address")
			log.WithField("pprofAddr", addr).Infof("Launch pprof runtime profiler")
			go func() {
				log.WithError(http.ListenAndServe(addr, nil)).Warnf("Could not start pprof")
			}()
		}

		return nil
	}

	app.Commands = commands
	app.Action = listenAction

	app.Run(os.Args)
}

func listenAction(ctx *cli.Context) error {
	config, ok := ctx.App.Metadata["config"].(*config.C)
	if !ok {
		return errors.Errorf("Could not load configuration")
	}

	addrStr := ctx.String("listen-address")
	port := ctx.Int("listen-port")

	listenAddr := fmt.Sprintf("%s:%d", addrStr, port)
	err := server.Listen(config, listenAddr)
	if err != nil {
		return errors.Wrap(err, "while running server")
	}

	return nil
}

package main

import (
	"fmt"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"

	"github.com/mailgun/holster/errors"
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
	}

	app.Before = func(ctx *cli.Context) error {
		verbose := ctx.Bool("verbose")
		if verbose {
			log.SetFormatter(&log.TextFormatter{})
			log.SetOutput(os.Stderr)
			log.SetLevel(log.DebugLevel)
			log.Debug("Verbose logging enabled")
		}

		return nil
	}

	app.Commands = commands
	app.Action = listenAction

	app.Run(os.Args)
}

func listenAction(ctx *cli.Context) error {
	addrStr := ctx.String("listen-address")
	port := ctx.Int("listen-port")

	listenAddr := fmt.Sprintf("%s:%d", addrStr, port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrap(err, "while opening listener socket")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "while accepting socket connection")
		}

		go handleConnection(conn)
	}
}

func handleConnection(client net.Conn) {
	for {
		data := make([]byte, 0)
		count, err := client.Read(data)
		if err != nil {
			log.WithError(err).Errorf("Failed reading data from connection, shutting down")
			err = client.Close()
			if err != nil {
				log.WithError(err).Errorf("Failed while shutting down connection")
			}

			return
		}

		fmt.Printf("received payload (%d bytes): %s\n", count, data)
	}
}

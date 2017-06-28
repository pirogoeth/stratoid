package server

import (
	"context"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/holster/errors"

	"github.com/pirogoeth/stratoid/config"
)

var serveContext context.Context
var serveCancel context.CancelFunc

func Listen(config *config.C, listenAddr string) error {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return errors.Wrap(err, "while trying to open tcp listener socket")
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "while opening listener socket")
	}

	serveContext, serveCancel = context.WithCancel(context.Background())
	worker := &ClientWorker{
		config:  config,
		clients: make(chan net.Conn, 1),
	}
	go worker.Start(context.WithCancel(serveContext))

	log.Infof("Starting accept loop on %s", listenAddr)

	for {
		listener.SetDeadline(time.Now().Add(1 * time.Second))
		conn, err := listener.Accept()
		if err != nil {
			if err, ok := err.(*net.OpError); ok && err.Timeout() {
				continue
			}

			serveCancel()
			return errors.Wrap(err, "while accepting socket connection")
		}

		log.Debugf("Accepting worker connection from: %s", conn.RemoteAddr().String())
		worker.clients <- conn
	}
}

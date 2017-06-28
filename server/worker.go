package server

import (
	"bufio"
	"context"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/pirogoeth/stratoid/config"
	"github.com/pirogoeth/stratoid/stratum"
)

const (
	eventWaitTimeout time.Duration = 10 * time.Second
)

type ClientWorker struct {
	config  *config.C
	clients chan net.Conn
}

type workerAssignment struct {
	pool        *config.Pool
	worker      *config.Worker
	loginParams stratum.Params
}

type minerClient struct {
	config     *config.C
	workQ      chan *workerAssignment
	clientSock net.Conn
	poolSock   net.Conn
	nextCmdId  int

	loopCancel   context.CancelFunc
	clientCancel context.CancelFunc
	poolCancel   context.CancelFunc
}

func (w *ClientWorker) Start(ctx context.Context, cancel context.CancelFunc) {
	defer close(w.clients)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Context closed, shutting down client worker")
			return
		case sock := <-w.clients:
			log.Debugf("Launching client event loop for %s", sock.RemoteAddr())
			client := &minerClient{
				config:     w.config,
				clientSock: sock,
				workQ:      make(chan *workerAssignment, 1),
			}
			go client.eventLoop(context.WithCancel(ctx))
		}
	}
}

func (m *minerClient) eventLoop(ctx context.Context, cancel context.CancelFunc) {
	defer close(m.workQ)
	defer cancel()

	m.loopCancel = cancel

	clientCtx, clientCancel := context.WithCancel(ctx)
	m.clientCancel = clientCancel
	go m.clientListener(clientCtx, clientCancel)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Stopping client event loop for %s", m.clientSock.RemoteAddr())
			m.poolCancel()
			m.clientCancel()
			return
		case assignment := <-m.workQ:
			log.WithFields(log.Fields{
				"worker":   assignment.worker.Username,
				"pool":     assignment.pool.Name,
				"username": assignment.pool.Username,
			}).Infof("Received pool assignment, initializing pool proxy for connected worker")

			poolConn, err := assignment.pool.OpenConnection()
			if err != nil {
				// XXX - Use backoff??
				log.WithError(err).Errorf("Could not connect to pool")
				return
			}

			poolLogin := stratum.CreateLoginRequest(
				assignment.pool.Username,
				assignment.pool.Password,
				assignment.loginParams,
			)

			err = poolLogin.Send(poolConn)
			if err != nil {
				log.WithError(err).Errorf("Failed while sending login packet to pool")
				return
			}

			log.Infof("Successfully wrote login request to pool (%#v)", poolLogin)

			m.poolSock = poolConn
			poolCtx, poolCancel := context.WithCancel(ctx)
			m.poolCancel = poolCancel
			go m.poolListener(poolCtx, poolCancel)
		}
	}
}

func (m *minerClient) clientListener(ctx context.Context, cancel context.CancelFunc) {
	reader := bufio.NewReader(m.clientSock)

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Context closed, closing client listener")
			m.clientSock.Close()
			return
		case <-time.After(time.Duration(1 * time.Millisecond)):
			break
		}

		data, pfx, err := reader.ReadLine()
		if pfx {
			log.Debugf("Payload received is chunked, assembling prefixes")
			for pfx {
				chunk, pfx, err := reader.ReadLine()
				if err != nil {
					log.WithError(err).Errorf("Failed reassembling line from socket connection")
					m.loopCancel()
					break
				}
				if pfx {
					data = append(data, chunk...)
				}
			}
		}
		if err != nil {
			log.WithError(err).Errorf("Failed reading data from connection, shutting down event loop")
			m.loopCancel()
			continue
		}

		if data == nil || len(data) == 0 {
			log.Debugf("Received empty payload from client")
			continue
		}

		log.Debugf("Received client payload (%d bytes): %s", len(data), data)

		// Client gives us data to process and or forward
		req, err := stratum.DecodeRequest(data)
		if err != nil {
			log.WithError(err).WithField("payload", data).Warnf("Could not decode payload")
			break
		}

		m.handleRequest(req)
	}
}

func (m *minerClient) poolListener(ctx context.Context, cancel context.CancelFunc) {
	reader := bufio.NewReader(m.poolSock)

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Context closed, closing pool listener")
			m.poolSock.Close()
			return
		case <-time.After(time.Duration(1 * time.Millisecond)):
			break
		}

		data, pfx, err := reader.ReadLine()
		if pfx {
			log.Debugf("Payload received is chunked, assembling prefixes")
			for pfx {
				chunk, pfx, err := reader.ReadLine()
				if err != nil {
					log.WithError(err).Errorf("Failed reassembling line from socket connection")
					m.loopCancel()
					break
				}
				if pfx {
					data = append(data, chunk...)
				}
			}
		}
		if err != nil {
			log.WithError(err).Errorf("Failed reading data from connection, closing socket")
			m.loopCancel()
			continue
		}

		if data == nil || len(data) == 0 {
			log.Warnf("Received empty payload from pool, closing connection")
			return
		}

		log.Debugf("Received pool payload (%d bytes): %s", len(data), data)

		// Pool gave us data for the client
		resp, err := stratum.DecodeResponse(data)
		if err != nil {
			log.WithError(err).WithField("payload", data).Warnf("Could not decode payload")
			break
		}

		m.handleResponse(resp)
	}
}

func (m *minerClient) handleRequest(req *stratum.Request) {
	if req.Method == "login" {
		workerName := req.Params["login"].(string)
		// XXX - Should we actually authorize a worker w/ password?
		// workerPass := req.Params["pass"].(string)

		// Get worker assignment
		worker, err := m.config.LookupWorker(workerName)
		if err != nil {
			log.WithError(err).Errorf("Could not look up worker login")
			// XXX - Send a "worker name invalid" response
		}

		pool, err := m.config.LookupPool(worker.Pool.Name)
		if err != nil {
			log.WithError(err).Errorf("Could not look up worker's assigned pool")
			// XXX - Send a "server error" response
		}

		assignment := &workerAssignment{
			worker:      worker,
			pool:        pool,
			loginParams: req.Params,
		}
		m.workQ <- assignment
	} else {
		if err := req.Send(m.poolSock); err != nil {
			log.WithError(err).Errorf("could not send response to pool")
		}
	}
}

func (m *minerClient) handleResponse(resp *stratum.Response) {
	// if !resp.IsCall() {
	// 	// XXX - Do neat shit here
	// }

	if err := resp.Send(m.clientSock); err != nil {
		log.WithError(err).Errorf("could not send response to client")
	}
}

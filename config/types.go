package config

import (
	"fmt"
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/holster/errors"
)

const (
	// ProtoStratum Protocol name for Stratum
	ProtoStratum = "stratum"

	// TransportTCP Transport name for TCP
	TransportTCP = "tcp"
)

type C struct {
	Pools   []*Pool   `toml:"pool"`
	Workers []*Worker `toml:"worker"`
}

// LookupPool searches for a single pool in the configuration
func (c *C) LookupPool(name string) (*Pool, error) {
	for _, pool := range c.Pools {
		if pool.Name == name {
			return pool, nil
		}
	}

	return nil, errors.WithContext{
		"poolName": name,
	}.Error("Could not find pool in configuration")
}

// LookupWorker searches for a worker configuration given the username
// for that worker.
func (c *C) LookupWorker(username string) (*Worker, error) {
	for _, worker := range c.Workers {
		if worker.Username == username {
			return worker, nil
		}
	}

	return nil, errors.WithContext{
		"workerUsername": username,
	}.Error("Could not find worker in configuration")
}

type Transport struct {
	Protocol  string
	Transport string
	Address   string
	Port      string
}

// AddrSpec builds the connection address with host and port.
func (t *Transport) AddrSpec() string {
	return fmt.Sprintf("%s:%s", t.Address, t.Port)
}

type TransportAddress string

// Parse parses a Transport Address (proto+transport://address:port) into
// a struct.
func (t TransportAddress) Parse() (*Transport, error) {
	addrStr := string(t)

	protoAry := strings.Split(addrStr, "+")
	protocol := protoAry[0]

	if protocol != ProtoStratum {
		return nil, errors.Errorf("Only '%s' protocol is supported, not '%s'", ProtoStratum, protocol)
	}

	transAry := strings.Split(protoAry[1], "://")
	transport := transAry[0]

	if transport != TransportTCP {
		return nil, errors.Errorf("Only '%s' transport is supported, not '%s'", TransportTCP, transport)
	}

	hostAry := strings.Split(transAry[1], ":")
	host := hostAry[0]
	port := hostAry[1]

	if strings.HasSuffix(port, "/") {
		port = port[:len(port)-1]
	}

	return &Transport{
		Protocol:  protocol,
		Transport: transport,
		Address:   host,
		Port:      port,
	}, nil
}

type Pool struct {
	Name     string
	Address  TransportAddress
	Username string
	Password string `default:"x"`
	Timeout  int    `default:"10"`
}

// OpenConnection creates a net.Conn to access the current pool.
func (p *Pool) OpenConnection() (net.Conn, error) {
	log.WithFields(log.Fields{
		"address":  p.Address,
		"poolName": p.Name,
		"timeout":  p.Timeout,
	}).Debugf("Attempting connection to pool")

	transport, err := p.Address.Parse()
	if err != nil {
		return nil, errors.Wrap(err, "while connecting to pool")
	}

	conn, err := net.Dial(transport.Transport, transport.AddrSpec())
	if err != nil {
		return nil, errors.WithContext{
			"transport": transport.Transport,
			"address":   transport.AddrSpec(),
		}.Wrap(err, "while dialing pool via transport")
	}

	return conn, nil
}

type Worker struct {
	Username string
	Pool     PoolRef

	PoolConn net.Conn
}

type PoolRef struct {
	Name string
}

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cloudmachinery/apps/http-userroles/server/store"
	"github.com/labstack/echo/v4"
	"github.com/swaggo/echo-swagger"
)

const (
	defaultListenerType = "unix"
	defaultSockAddr     = "/tmp/http-userroles.sock"
	defaultConnString   = "boltdb://./userroles.boltdb"
)

type config struct {
	addr         string
	listenerType string
	connString   string
	keyPath      string
	certPath     string
	TLSAddr      string
}

func main() {
	var c config
	flag.StringVar(&c.listenerType, "listener-type", defaultListenerType, "Listener type")
	flag.StringVar(&c.addr, "addr", defaultSockAddr, "Socket address")
	flag.StringVar(&c.connString, "conn-string", defaultConnString, "Socket address")
	flag.StringVar(&c.keyPath, "key", "", "Key path")
	flag.StringVar(&c.certPath, "cert", "", "Cert path")
	flag.StringVar(&c.TLSAddr, "tls-addr", "", "TLS address")
	flag.Parse()

	if c.listenerType == "unix" {
		_ = os.Remove(c.addr)
	}

	l, err := net.Listen(c.listenerType, c.addr)
	failOnError(err, "unable to listen to socket")
	defer l.Close()

	e := echo.New()
	e.Listener = l
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	store, err := getStore(c.connString)
	failOnError(err, "unable to create store")
	defer store.Close(context.Background())

	h := NewHandler(store)

	h.RegisterRoutes(e)

	if c.TLSAddr != "" {
		err = e.StartTLS(c.TLSAddr, c.certPath, c.keyPath)
	} else {
		err = e.Start("")
	}
	failOnError(err, "unable to start")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getStore(connString string) (store.Store, error) {
	if strings.HasPrefix(connString, "postgres://") {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return store.NewPostgresStore(ctx, connString)
	}

	if strings.HasPrefix(connString, "boltdb://") {
		connString = strings.TrimPrefix(connString, "boltdb://")
		return store.NewBoltStore(connString)
	}

	return nil, errors.New("unknown connection string. Only boltdb://* and postgres://* are supported")
}

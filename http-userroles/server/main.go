package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/swaggo/echo-swagger"
)

const (
	unixSockAddr = "/tmp/http-userroles.sock"
	dbPath       = "./userroles.boltdb"
)

func main() {
	var addr string
	var listenerType string
	flag.StringVar(&listenerType, "l", "unix", "Listener type")
	flag.StringVar(&addr, "a", "/tmp/http-userroles.sock", "Socket address")
	flag.Parse()

	if listenerType == "unix" {
		_ = os.Remove(unixSockAddr)
	}

	l, err := net.Listen(listenerType, addr)
	failOnError(err, "unable to listen to socket")
	defer l.Close()

	e := echo.New()
	e.Listener = l
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	store, err := NewStore(dbPath)
	failOnError(err, "unable to create store")
	defer store.Close()

	h := NewHandler(store)

	h.RegisterRoutes(e)

	err = e.Start("")
	failOnError(err, "unable to start")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

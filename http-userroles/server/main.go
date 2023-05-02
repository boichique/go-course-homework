package main

import (
	"log"
	"net"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	sockAddr = "/tmp/http-userroles.sock"
	dbPath   = "./userroles.boltdb"
)

func main() {
	_ = os.Remove(sockAddr)

	l, err := net.Listen("unix", sockAddr)
	failOnError(err, "unable to listen to socket")

	defer l.Close()

	e := echo.New()
	e.Listener = l

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

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/cloudmachinery/apps/tcp-guessgame/message"
)

func main() {
	addr, err := getServerAddr()
	failOnError(err, "cannot resolve server address")

	// DialTCP
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalf("cannot connect to server: %v", err)
	}

	// Send "start" message and get the range
	err = message.Write(conn, message.Start)
	if err != nil {
		log.Fatalf("cannot send start message: %v", err)
	}

	mes, err := message.Read(conn)
	if err != nil {
		log.Fatalf("cannot get message from server: %v", err)
	}

	var min, max int
	n, err := fmt.Sscanf(mes, message.MinMaxFormat, &min, &max)
	if err != nil {
		log.Fatalf("cannot get range: %v", err)
	}
	if n != 2 {
		log.Fatalf("scanned %d elements, expected 2", n)
	}

	fmt.Printf("min: %d, max: %d\n", min, max)

	for min <= max {
		//  use binary search to guess the number
		//  guess in the middle of the range
		guess := min + (max-min)/2
		fmt.Printf("guessing %d\n", guess)

		//  send the guess to the server
		err = message.Write(conn, strconv.Itoa(guess))
		if err != nil {
			log.Fatalf("cannot send the guess to the server: %s", err)
		}

		//  read the response from the server and log it
		res, err := message.Read(conn)
		if err != nil {
			log.Fatalf("cannot read the response from the server: %s", err)
		}
		fmt.Println(res)

		switch res {
		case message.Higher:
			min = guess + 1
		case message.Lower:
			max = guess - 1
		case message.Correct:
			return
		}
	}
	log.Fatal("game ended unexpectedly")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getServerAddr() (*net.TCPAddr, error) {
	var (
		host string
		port int
	)
	flag.StringVar(&host, "host", "localhost", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
}

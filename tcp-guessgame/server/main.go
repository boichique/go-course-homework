package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cloudmachinery/apps/tcp-guessgame/message"
)

const (
	RandSeedEnvVar = "RAND_SEED"
)

func main() {
	cfg, err := getConfig()
	failOnError(err, "cannot get config")

	// Listen for fmt.Sprintf(":%d", cfg.Port) tcp port
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	failOnError(err, "cannot listen")
	defer l.Close()

	log.Printf("listening on :%d\n", cfg.Port)

	guesser := Guesser{
		Min: cfg.Min,
		Max: cfg.Max,
		Gen: rand.New(rand.NewSource(cfg.Seed)),
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("cannot accept: %s\n", err)
			continue
		}

		go func() {
			defer conn.Close()
			err := guesser.Play(conn)
			if err != nil {
				log.Printf("error playing game: %s", err)
			}
		}()
	}
}

type Guesser struct {
	Min, Max int
	Gen      *rand.Rand
}

func (g *Guesser) Play(conn net.Conn) error {
	// read message. it should be start, otherwise return error
	mes, err := message.Read(conn)
	if err != nil {
		return fmt.Errorf("cannot read message: %w", err)
	}
	if mes != message.Start {
		return fmt.Errorf("game is not started")
	}

	// write configured min and max
	minMaxRange := fmt.Sprintf(message.MinMaxFormat, g.Min, g.Max)
	err = message.Write(conn, minMaxRange)
	if err != nil {
		return fmt.Errorf("cannot send min and max to the client: %w", err)
	}

	// make a guess and log it
	guess := g.Gen.Intn(g.Max-g.Min+1) + g.Min
	log.Printf("guessed %d\n", guess)

	isGuessed := false
	for !isGuessed {
		//read client guess
		mes, err = message.Read(conn)
		if err != nil {
			return fmt.Errorf("cannot read message: %w", err)
		}

		//convert to int
		num, err := strconv.Atoi(string(mes))
		if err != nil {
			return fmt.Errorf("cannot convert message to int: %w", err)
		}

		//use switch to compare numbers and return appropriate message
		//in case of correct guess end the game and return nil
		switch {
		case num > guess:
			err = message.Write(conn, message.Lower)
			if err != nil {
				return fmt.Errorf("cannot send \"lower\" to the client: %w", err)
			}
		case num < guess:
			err = message.Write(conn, message.Higher)
			if err != nil {
				return fmt.Errorf("cannot send \"higher\" to the client: %w", err)
			}
		default:
			err = message.Write(conn, message.Correct)
			if err != nil {
				return fmt.Errorf("cannot send \"correct\" to the client: %w", err)
			}
			isGuessed = true
		}
	}
	return nil
}

type config struct {
	Port int
	Min  int
	Max  int
	Seed int64
}

func getConfig() (config, error) {
	var c config
	flag.IntVar(&c.Port, "port", 8080, "port to listen on")
	flag.IntVar(&c.Min, "min", 0, "minimum number to guess")
	flag.IntVar(&c.Max, "max", 100, "maximum number to guess")
	flag.Parse()

	if c.Min >= c.Max {
		return c, fmt.Errorf("min must be less than max")
	}

	seedEnv, ok := os.LookupEnv(RandSeedEnvVar)
	if ok {
		seed, env := strconv.Atoi(seedEnv)
		if env != nil {
			return c, fmt.Errorf("invalid seed value: %s", seedEnv)
		}

		c.Seed = int64(seed)
	} else {
		c.Seed = time.Now().UnixNano()
	}

	return c, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

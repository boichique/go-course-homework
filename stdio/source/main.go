package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	seedEnvVar = "RAND_SEED"
)

func main() {
	// flag vars
	var interval time.Duration
	var duration time.Duration
	var from int
	var to int

	// define and parse flags
	flag.DurationVar(&interval, "i", 100*time.Millisecond, "Interval for generation")
	flag.DurationVar(&duration, "d", 10*time.Second, "Duration of generation")
	flag.IntVar(&from, "f", 0, "Min range index")
	flag.IntVar(&to, "t", 100, "Max range index")

	flag.Parse()

	// generator, err := getRandGenerator()
	generator, err := getRandGenerator()
	if err != nil {
		log.Fatal(err)
	}

	// for ; ; {
	//   generate metrics
	// {
	for t := time.Now(); time.Since(t) < duration; time.Sleep(interval) {
		fmt.Printf("metric: %d\n", (generator.Intn(to-from) + from))
	}
}

func getRandGenerator() (*rand.Rand, error) {
	seed, ok := os.LookupEnv(seedEnvVar)
	if !ok {
		return nil, fmt.Errorf("missing environment variable: %s", seedEnvVar)
	}

	seedInt, err := strconv.Atoi(seed)
	if err != nil {
		return nil, fmt.Errorf("can't convert seed into int: %s", err)
	}

	generator := rand.New(rand.NewSource(int64(seedInt)))
	// get seed from env var and create generator
	return generator, nil
}

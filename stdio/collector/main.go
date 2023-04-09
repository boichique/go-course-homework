package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
)

const (
	noMin = math.MaxInt
	noMax = math.MinInt
)

func main() {
	// create scanner on top of stdin
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// read lines in a loop
	min, max := noMin, noMax
	sum := 0
	counter := 0
	rangeChanged := false

	// handler scanner error
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	for scanner.Scan() {
		var metric int //   parse line
		line := scanner.Text()
		_, err := fmt.Sscanf(line, "metric: %d", &metric)
		if err != nil {
			log.Fatal(err)
		}

		if metric < min { //   update stats
			min = metric
			rangeChanged = true
		}
		if metric > max {
			max = metric
			rangeChanged = true
		}
		counter++
		sum += metric

		if rangeChanged { //   update range and print if changed
			fmt.Fprintf(os.Stdout, "New range: [%d:%d]\n", min, max)
			rangeChanged = false
		}
	}

	// print final stats
	fmt.Fprintf(os.Stdout, "Stats: metrics collected: %d in range [%d:%d] with avg: %s\n", counter, min, max, fmt.Sprintf("%.2f", float64(sum)/float64(counter)))

}

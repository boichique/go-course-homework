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
	r := fmt.Sprintf("[%d:%d]", min, max)

	for scanner.Scan() {
		var metric int //   parse line
		line := scanner.Text()
		_, err := fmt.Sscanf(line, "metric: %d", &metric)
		if err != nil {
			log.Fatal(err)
		}

		if metric < min { //   update stats
			min = metric
		}
		if metric > max {
			max = metric
		}
		counter++
		sum += metric

		if metric == max || metric == min { //   update range and print if changed
			r = fmt.Sprintf("[%d:%d]", min, max)
			fmt.Fprintf(os.Stdout, "New range: %s\n", r)
		}

		// handler scanner error
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}

	// print final stats
	avg := fmt.Sprintf("%.2f", float64(sum)/float64(counter))
	fmt.Fprintf(os.Stdout, "Stats: metrics collected: %d in range %s with avg: %s\n", counter, r, avg)

}

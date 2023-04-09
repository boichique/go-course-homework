package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Stat struct {
	Word  string
	Count int
}

func main() {
	// parse flags
	var in string
	var out string
	var limit int
	var minLen int

	flag.StringVar(&in, "in", "example.txt", "Path to input file")
	flag.StringVar(&out, "out", "output.txt", "Path to output file")
	flag.IntVar(&limit, "limit", 10, "Limit for word pairs in output file")
	flag.IntVar(&minLen, "min-length", 5, "Min length for word length in output file")

	flag.Parse()

	// read file
	readFile, err := os.Open(in)
	if err != nil {
		fmt.Println("Unable to open file:", err)
		log.Fatal(err)
	}
	defer readFile.Close()

	// index words
	freq := map[string]int{}

	inData, err := io.ReadAll(readFile)

	if err != nil {
		fmt.Println("Unable to read file:", err)
		log.Fatal(err)
	}

	words := strings.FieldsFunc(string(inData), func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	for _, word := range words {
		freq[strings.ToLower(word)]++
	}

	// sort pairs
	var wc []Stat
	for k, v := range freq {
		if len(k) >= minLen {
			wc = append(wc, Stat{k, v})
		}
	}

	sort.Slice(wc, func(i, j int) bool {
		if wc[i].Count == wc[j].Count {
			return wc[i].Word < wc[j].Word
		}
		return wc[i].Count > wc[j].Count
	})

	// limit pairs
	var outData strings.Builder
	for i := 0; i < len(wc) && limit > 0; i, limit = i+1, limit-1 {
		outData.WriteString(fmt.Sprintf("%s: %d\n", wc[i].Word, wc[i].Count))
	}

	// write file
	writeFile, err := os.Create(out)
	if err != nil {
		fmt.Println("Unable to create file:", err)
		log.Fatal(err)
	}
	defer writeFile.Close()
	writeFile.Write([]byte(outData.String()))
}

package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
)

func main() {
	path, err := getPath()
	if err != nil {
		log.Fatal(fmt.Errorf("get path: %w", err))
	}

	// file/dir does not exist?
	// log.Fatal(fmt.Errorf("check path %q: %w", path, err))

	if err = printHashes(os.DirFS(path), os.Stdout, "."); err != nil {
		log.Fatal(err)
	}
}

func printHashes(f fs.FS, w io.Writer, path string) error {
	return fmt.Errorf("smth")
	// iterate through files/dirs in path
	//   if dir -> skip
	//   open file and calc hash
	//   print hash and path in that format: "sha256:%s %s\n", hash, path
}

func getPath() (string, error) {
	return "err", fmt.Errorf("smth")
	// args := os.Args[1:]

	// return "", fmt.Errorf("too many arguments: usage: filehashes [path]")
}

func calcHash(f fs.File) (string, error) {
	return "err", fmt.Errorf("smth")
	// create sha256 hasher
	// copy file content to hasher
	// return hex encoded hash
}

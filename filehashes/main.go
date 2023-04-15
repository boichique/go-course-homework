package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	path, err := getPath()
	if err != nil {
		log.Fatal(fmt.Errorf("check path %q: %w", path, err))
	}

	// file/dir does not exist?
	// log.Fatal(fmt.Errorf("check path %q: %w", path, err))

	if err = printHashes(os.DirFS(path), w, "."); err != nil {
		log.Fatal(err)
	}
}

func printHashes(f fs.FS, w io.Writer, path string) error {
	// iterate through files/dirs in path
	err := fs.WalkDir(f, path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		//   if dir -> skip
		if d.IsDir() {
			return nil
		}
		//   open file and calc hash
		file, err := f.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", path, err)
		}
		defer file.Close()

		hash, err := calcHash(file)
		if err != nil {
			return fmt.Errorf("unable to calc hash for file %s: %w", path, err)
		}
		//   print hash and path in that format: "sha256:%s %s\n", hash, path

		fmt.Fprintf(w, "sha256:%s %s\n", hash, path)

		return nil
	})

	return err
}

func getPath() (string, error) {
	// args := os.Args[1:]
	args := os.Args[1:]
	if len(args) == 0 {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal()
		}
		return path, nil
	}

	// return "", fmt.Errorf("too many arguments: usage: filehashes [path]")
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments: usage: filehashes [path]")
	}
	path := args[0]
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("path %q does not exist", path)
	}
	return path, nil
}

func calcHash(f fs.File) (string, error) {
	// create sha256 hasher
	hasher := sha256.New()

	// copy file content to hasher
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	// return hex encoded hash
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

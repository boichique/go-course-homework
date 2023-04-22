package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

type CountMethod func(path string) (words int, err error)
type Modes map[string]CountMethod

func (m Modes) All() []string {
	var modes []string
	for mode := range m {
		modes = append(modes, mode)
	}
	return modes
}

func (m Modes) IsAllowed(mode string) bool {
	_, ok := m[mode]
	return ok
}

var allowedModes = Modes{
	"sequential":       calculateSequentially,
	"parallel":         calculateParallel,
	"limited-parallel": calculateLimitedParallel,
}

type config struct {
	path       string
	mode       string
	cpuProfile string
	trace      string
}

func getConfig() (config, error) {
	var cfg config

	flag.StringVar(&cfg.mode, "mode", "sequential", fmt.Sprintf("Mode to run the program in %s", allowedModes.All()))
	flag.StringVar(&cfg.path, "path", ".", "path to the directory to process")
	flag.StringVar(&cfg.cpuProfile, "cpu-profile", "", "write cpu profile to file")
	flag.StringVar(&cfg.trace, "trace", "", "write trace to file")
	flag.Parse()

	if !allowedModes.IsAllowed(cfg.mode) {
		return cfg, fmt.Errorf("invalid mode %q, allowed modes: %s", cfg.mode, allowedModes.All())
	}

	return cfg, nil
}

func main() {
	cfg, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("cores: %d, config: %+v\n", runtime.NumCPU(), cfg)

	if cfg.cpuProfile != "" {
		f, err := os.Create(cfg.cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("failed to start writing cpu profile to file %q: %v", cfg.cpuProfile, err)
		} else {
			defer pprof.StopCPUProfile()
		}
	}

	if cfg.trace != "" {
		f, err := os.Create(cfg.trace)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := trace.Start(f); err != nil {
			fmt.Printf("failed to start writing traces to file %q: %v", cfg.trace, err)
		} else {
			defer trace.Stop()
		}
	}

	method := allowedModes[cfg.mode]

	var totalCount int
	totalCount, err = method(cfg.path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total words count: %d\n", totalCount)
}

func calculateParallel(path string) (int, error) {
	// Implement the same logic as in calculateSequential, but count words for each file in a separate goroutine
	var wg sync.WaitGroup
	var totalCount atomic.Int64

	err := traverseThroughAllFiles(os.DirFS(path), func(f fs.FS, path string) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := countWords(f, path)
			if err != nil {
				log.Printf("count words in %q: %v", path, err)
				return
			}
			totalCount.Add(int64(count))
		}()
		return nil
	})

	wg.Wait()
	return int(totalCount.Load()), err
}

func calculateLimitedParallel(path string) (int, error) {
	// Implement the same logic as in calculateParallel, but process each path in separate workers. Amount of workers should be equal to amount of CPU cores.
	// Use channels
	goroutines := runtime.NumCPU()
	var wg sync.WaitGroup
	var totalCount atomic.Int64
	ch := make(chan string)
	f := os.DirFS(path)
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for path := range ch {
				count, err := countWords(f, path)
				if err != nil {
					log.Printf("count words in %q: %v", path, err)
					continue
				}
				totalCount.Add(int64(count))
			}
		}()
	}

	err := traverseThroughAllFiles(os.DirFS(path), func(f fs.FS, path string) error {
		ch <- path
		return nil
	})

	close(ch)

	if err != nil {
    return 0, err
	}

	wg.Wait()

	return int(totalCount.Load()), nil
}

func calculateSequentially(path string) (int, error) {
	var totalCount int

	err := traverseThroughAllFiles(os.DirFS(path), func(f fs.FS, path string) error {
		count, err := countWords(f, path)
		if err != nil {
			log.Printf("count words in %q: %v", path, err)
			return nil
		}

		totalCount += count
		return nil
	})

	return totalCount, err
}

func traverseThroughAllFiles(f fs.FS, fn func(f fs.FS, path string) error) error {
	return fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if d.IsDir() {
			return nil
		}

		return fn(f, path)
	})
}

func countWords(f fs.FS, path string) (int, error) {
	file, err := f.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return len(words), nil
}

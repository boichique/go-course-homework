package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := getConfig()
	failIfErr(err)

	// Subscribe to os.Interrupt and os.Kill signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cancel the context when the program receives a signal
	go func() {
		sig := <-sigChan
		log.Printf("received signal: %s, cancelling context", sig)
		cancel()
	}()

	err = repeat(ctx, cfg)
	failIfErr(err)
}

func repeat(ctx context.Context, cfg config) error {
	// run once before starting the loop
	if err := runCmd(cfg.cmd, cfg.args...); err != nil {
		return err
	}

	// run the command every cfg.interval using a ticker
	// stop repeating when the context is cancelled
	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled")
			return nil
		case <-ticker.C:
			if err := runCmd(cfg.cmd, cfg.args...); err != nil {
				log.Printf("error running command: %s", err)
			}
		}
	}
}

func runCmd(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	out, err := c.Output()
	if err != nil {
		return err
	}
	os.Stdout.Write(out)
	return nil
}

func failIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	interval time.Duration
	cmd      string
	args     []string
}

func getConfig() (config, error) {
	var cfg config

	flag.DurationVar(&cfg.interval, "interval", 2*time.Second, "interval to wait between runs")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return cfg, fmt.Errorf("no command specified. usage: watchcmd [flags] <cmd> [args]")
	}

	cfg.cmd = args[0]
	if len(args) > 1 {
		cfg.args = args[1:]
	}
	return cfg, nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Dbg             bool          `long:"dbg" env:"DEBUG" description:"show debug info"`
	Version         bool          `short:"v" description:"Show version and exit"`
	AttemptsAllowed int           `short:"a" description:"Number of failed attempts allowed before reboot" default:"3"`
	Address         string        `long:"address" description:"Address to check" default:"https://www.google.com"`
	Interval        time.Duration `short:"i" long:"interval" env:"INTERVAL" default:"15m" description:"Interval between checks"`
	RetryInterval   time.Duration `short:"r" long:"retry-interval" env:"RETRY_INTERVAL" default:"5m" description:"Interval between checks after failed attempt"`
}

var version = "undefined" // version is set during build

func main() {
	// Parsing cmd parameters
	var opts Options
	p := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	// Logger setup
	logOpts := []lgr.Option{
		lgr.LevelBraces,
		lgr.StackTraceOnError,
	}
	if opts.Dbg {
		logOpts = append(logOpts, lgr.Debug)
	}
	lgr.SetupStdLogger(logOpts...)

	// Graceful termination
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Shutdown signal received\n*********************************")
		cancel()
	}()

	defer func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic: %+v", x)
		}
	}()

	if opts.Version {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	worker(ctx, opts)
}

func worker(ctx context.Context, opts Options) {

	log.Printf("[INFO] Worker started with options: \n* Interval: %s \n* Retry Interval: %s \n* Attempts Allowed: %d\n* Check Address: %s \nversion: %s",
		opts.Interval.String(), opts.RetryInterval.String(), opts.AttemptsAllowed, opts.Address, version)

	interval := opts.Interval

	failedAttempts := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Worker stopped: %s", ctx.Err())
			return
		case <-time.After(interval):
			log.Printf("[INFO] Interval passed: %s", interval)

			// Network check
			err := checkNetwork(opts.Address)
			if err == nil {
				log.Printf("[INFO] Network check passed")
				failedAttempts = 0
				interval = opts.Interval
				continue
			}

			failedAttempts++
			interval = opts.RetryInterval
			log.Printf("[INFO] Network check failed: %s", err)
			log.Printf("[DEBUG] Failed attempts: %d", failedAttempts)

			// First failed GET - attempt to fix the issue by restarting systemd-resolved
			if failedAttempts == 1 {
				exec.Command("systemctl", "restart", "systemd-resolved").Run()
				log.Printf("[INFO] Restarted systemd-resolved")
			}

			if failedAttempts > opts.AttemptsAllowed {
				log.Printf("[INFO] Network check failed %d times, rebooting system", failedAttempts)
				if err := reboot(); err != nil {
					log.Printf("[ERROR] Failed to reboot: %e", err)
				}
			}
		}

	}
}

func checkNetwork(addr string) error {
	// checking network status (if addr is reachable)
	// this should fail if there is no network connection or dns resolution is not working
	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}
	_, err := httpClient.Get(addr)
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", addr, err)
	}

	return err
}

// reboot reboots the system
func reboot() error {
	log.Printf("[INFO] !!! REBOOT CALLED !!!")
	// return nil
	syscall.Sync()
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

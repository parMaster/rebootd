package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Dbg             bool          `long:"dbg" env:"DEBUG" description:"show debug info"`
	Version         bool          `short:"v" description:"Show version and exit"`
	AttemptsAllowed int           `short:"a" description:"Number of failed attempts allowed before reboot" default:"5"`
	Address         string        `long:"address" description:"Address list to check - comma separated" default:"https://www.google.com,https://www.cloudflare.com,https://www.amazon.com"`
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
			err := testConnection(opts.Address)
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

			// First failed check - attempt to fix the issue by restarting systemd-resolved and retry
			if failedAttempts == 1 {
				exec.Command("systemctl", "restart", "systemd-resolved").Run()
				log.Printf("[INFO] Restarted systemd-resolved")
				time.Sleep(5 * time.Second)
				exec.Command("systemctl", "restart", "NetworkManager").Run()
				log.Printf("[INFO] Restarted NetworkManager")
				continue
			}

			if failedAttempts >= opts.AttemptsAllowed {
				log.Printf("[INFO] Network check failed %d times, rebooting system", failedAttempts)
				if err := reboot(); err != nil {
					log.Printf("[ERROR] Failed to reboot: %e", err)
				}
			}
		}

	}
}

// testConnection checks if the connection is available to the provided address slice (comma separated).
// It returns nil if at least one of the addresses is reachable
func testConnection(addr string) error {
	if addr == "" {
		return fmt.Errorf("empty address")
	}

	urls := strings.Split(addr, ",")
	for _, url := range urls {
		url = strings.TrimSpace(url)
		err := getWithTimeout(url)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("all connection tests failed: %s", addr)
}

// getWithTimeout performs a GET request with a timeout
func getWithTimeout(addr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", addr, nil)
	_, err := http.DefaultClient.Do(req)
	return err
}

// reboot reboots the system (Linux only) using syscall.Reboot.
// Won't work on Windows, MacOS, etc. Won't even compile.
func reboot() error {
	log.Printf("[INFO] !!! REBOOT CALLED !!!")
	syscall.Sync()
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	// return nil
}

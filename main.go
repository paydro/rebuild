package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"

	flag "github.com/spf13/pflag"
)

const Version = "0.0.1"

var buildCmd string
var watch string
var excludes []string
var version bool

func init() {
	flag.Usage = usage
	flag.ErrHelp = errors.New("")
	flag.StringVarP(&buildCmd, "build", "b", "", "build command")
	flag.StringVarP(&watch, "watch", "w", ".", "directory path to watch for changes")
	flag.StringSliceVarP(&excludes, "exclude", "e", []string{}, "directories to exclude")
	flag.BoolVarP(&version, "version", "v", false, "show version")
}

func usage() {
	cmd := path.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "    %s [OPTIONS] -- COMMAND\n", cmd)
	fmt.Fprintf(os.Stderr, "e.g.\n")
	fmt.Fprintf(os.Stderr, "    %s --build 'go build -o mytool .' -- mytool\n", cmd)
	fmt.Fprintf(os.Stderr, "    %s --exclude dir1 --exclude dir2 -- mytool\n", cmd)
	fmt.Fprintf(os.Stderr, "    %s --watch assets/ -- npm build\n\n", cmd)
	fmt.Fprintf(os.Stderr, "FLAGS:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("v%s\n", Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Printf("Missing command to execute!\n\n")
		flag.Usage()
		os.Exit(1)
	}

	runCmd := flag.Args()
	ctx, cancel := context.WithCancel(context.Background())

	eventNotifier := make(chan struct{})
	var wg sync.WaitGroup

	absPath, err := filepath.Abs(watch)
	if err != nil {
		log.Printf("Could not get absolute path of %s\n", watch)
		log.Fatalf("Please provide a different --watch argument.")
	}

	// start process manager
	process := NewPIDManager(runCmd, buildCmd)
	process.Start()
	wg.Add(1)
	go process.Listen(ctx, &wg, eventNotifier)

	// TODO add ability for watcher to signal a failure (i.e., the Start()
	// function died or something so main can properly shut down pidmanager or
	// vice versa)
	watcher := NewFileWatcher(absPath, excludes)
	wg.Add(1)
	go watcher.Start(ctx, &wg, eventNotifier)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	sig := <-sigs

	fmt.Printf("\n-----------------\n")
	fmt.Printf("Signal caught: %s\n", sig)
	fmt.Printf("Canceling go routines ...\n")
	fmt.Printf("-----------------\n")
	cancel()

	fmt.Println("Waiting for go routines to finish up")
	wg.Wait()

	fmt.Println("See ya next time!")
}

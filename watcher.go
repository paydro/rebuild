package main

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// minDelay timeout is dependent on the file that's being written
// and the plugins I've stored in VIM. For instance, saving
// main will call goimports. From what it looks like VIM will
// write the file twice.
const minDelay = 600 * time.Millisecond

type fileWatcher struct {
	dir      string
	excludes []string
	watcher  *fsnotify.Watcher
}

func NewFileWatcher(dir string, excludes []string) *fileWatcher {
	return &fileWatcher{
		dir:      dir,
		excludes: excludes,
	}
}

func (f *fileWatcher) Start(ctx context.Context, wg *sync.WaitGroup, notify chan struct{}) {
	defer wg.Done()

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Could not start file watcher: %s\n", err)
	}
	f.watcher = w
	defer f.watcher.Close()

	stat, err := os.Stat(f.dir)
	if err != nil {
		log.Fatalf("Could not stat() %s\n", watch)
		log.Fatalf("Please provide a different --watch argument.")
	}

	if stat.IsDir() {
		err = addDirectories(f.watcher, f.dir, f.excludes)
		if err != nil {
			log.Fatalf("Could not recursively add all directories under %s\n", watch)
			log.Fatalf("Please provide a different --watch argument.")
		}
	} else {
		f.watcher.Add(f.dir)
	}

	// Counter to let us know if we any events worth notifying on.
	eventCount := 0

	// Flag to notify. This is turned off when an event is sent on notify
	// to allow for the other process manipulate the filesystem without
	// this go routine continually emitting events.
	//
	// TODO Think about moving this flag to pidManager
	// TODO Another reason this flag exists is due to us watching 'wikitravel'
	// being built. We could ignore this run command if we wanted to and we
	// wouldn't have to worry about starting/stopping event emitting. That
	// said, it's not obvious if the runCmd is actually a binary or not. So, it
	// might not need to.
	notifyOnEvents := true

	for {
		select {
		case <-ctx.Done():
			log.Println("Finishing fswatcher")
			return

		case _, ok := <-f.watcher.Events:
			if !ok {
				log.Fatalf("Could not read event from filesystem.")
			}

			if notifyOnEvents {
				eventCount += 1
			}

		case err, ok := <-f.watcher.Errors:
			if !ok {
				log.Fatalf("Error reading event from filesystem: %s", err)
			}

			log.Println("error:", err)

		case <-notify:
			notifyOnEvents = true

		case <-time.After(minDelay):
			if notifyOnEvents && eventCount > 0 {
				// Disable notify until the other end completes
				notifyOnEvents = false
				notify <- struct{}{}
			}

			eventCount = 0
		}
	}
}

func addDirectories(watcher *fsnotify.Watcher, dir string, excludes []string) error {
	return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// log.Printf("Error walking path: %s. Skipping\n", p)
			return nil
		}

		// We only watch directories
		if !info.IsDir() {
			return nil
		}

		// Excludes
		for _, exclude := range excludes {
			if strings.HasPrefix(p, path.Join(dir, exclude)) {
				return filepath.SkipDir
			}
		}

		watcher.Add(p)

		return nil
	})
}

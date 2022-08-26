package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Info("event:", event)
				LoadConf()
				Logger.Info("conf change to:", Conf.String())
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				Info("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

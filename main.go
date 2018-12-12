package main

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

var startTime = time.Now()
var totalBytes int64
var numDownloaded int64
var downloadGroup sync.WaitGroup
var helperGroup sync.WaitGroup
var exitRequested int32

type Track struct {
	Artist   string   `json:"artist"`
	Title    string   `json:"title"`
	Album    string   `json:"album"`
	Genres   []string `json:"genre"`
	Download string   `json:"download"`
}

func main() {
	if err := parseArgs(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logrus.Info("Starting Free Music Archive Scraper")
	logrus.Info("  https://github.com/terorie/fma-scraper/")

	jobs := make(chan Track, 2 * *pageSize)
	results := make(chan Track, 2 * *pageSize)

	c, cancel := context.WithCancel(context.Background())

	go listenCtrlC(cancel)
	go stats()

	// Start logger
	helperGroup.Add(1)
	go logger(results)

	// Start downloaderss
	downloadGroup.Add(int(*concurrency))
	for i := 0; i < int(*concurrency); i++ {
		go downloader(jobs, results)
	}

	// Start meta grabber
	page := *minPage
	for {
		if atomic.LoadInt32(&exitRequested) != 0 {
			break
		}

		err := backoff.Retry(func() error {
			err := list(c, jobs, *genre, int(page))
			if err != nil {
				logrus.WithError(err).
					Errorf("Failed visiting page %d", page)
			}
			return err
		}, backoff.NewExponentialBackOff())

		if err != nil {
			logrus.Fatal(err)
		}

		page++
	}

	// Shutdown
	close(jobs)
	downloadGroup.Wait()
	close(results)
	helperGroup.Wait()

	total := atomic.LoadInt64(&totalBytes)
	dur := time.Since(startTime).Seconds()

	logrus.WithFields(logrus.Fields{
		"total_bytes": total,
		"dur": dur,
		"avg_rate": float64(total) / dur,
	}).Info("Stats")
}

func listenCtrlC(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	atomic.StoreInt32(&exitRequested, 1)
	cancel()
	fmt.Fprintln(os.Stderr, "\nWaiting for downloads to finish...")
	fmt.Fprintln(os.Stderr, "Press ^C again to exit instantly.")
	<-c
	fmt.Fprintln(os.Stderr, "\nKilled!")
	os.Exit(255)
}

func stats() {
	for range time.NewTicker(time.Second).C {
		total := atomic.LoadInt64(&totalBytes)
		dur := time.Since(startTime).Seconds()

		logrus.WithFields(logrus.Fields{
			"tracks": numDownloaded,
			"total_bytes": totalBytes,
			"avg_rate": fmt.Sprintf("%.0f", float64(total) / dur),
		}).Info("Stats")
	}
}

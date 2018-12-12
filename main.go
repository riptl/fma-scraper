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

var availableGenres = [...]string{
	"Blues",
	"Classical",
	"Country",
	"Electronic",
	"Experimental",
	"Folk",
	"Hip-Hop",
	"Instrumental",
	"International",
	"Jazz",
	"Novelty",
	"Old-Time__Historic",
	"Pop",
	"Rock",
	"Soul-RB",
	"Spoken",
}

var startTime = time.Now()
var totalDownloaded int64
var downloadGroup sync.WaitGroup
var helperGroup sync.WaitGroup
var exitRequested int32

const pageSize = 2

type Track struct {
	Artist   string   `json:"artist"`
	Title    string   `json:"title"`
	Album    string   `json:"album"`
	Genres   []string `json:"genre"`
	Download string   `json:"download"`
}

func main() {
	parseArgs()

	jobs := make(chan Track, 2 * pageSize)
	results := make(chan Track, 2 * pageSize)

	c, cancel := context.WithCancel(context.Background())

	go listenCtrlC(cancel)

	// Start logger
	helperGroup.Add(1)
	go logger(results)

	// Start downloaderss
	downloadGroup.Add(*concurrency)
	for i := 0; i < *concurrency; i++ {
		go downloader(jobs, results)
	}

	// Start meta grabber
	page := *minPage
	for {
		if atomic.LoadInt32(&exitRequested) != 0 {
			break
		}

		err := backoff.Retry(func() error {
			err := list(c, jobs, *genre, page)
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

	total := atomic.LoadInt64(&totalDownloaded)
	dur := time.Since(startTime).Seconds()

	logrus.WithFields(logrus.Fields{
		"total": total,
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

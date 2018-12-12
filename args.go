package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

var parser = argparse.NewParser("free-music-archive-scraper",
	"Scraper for https://freemusicarchive.org/")

var genre = parser.Selector("g", "genre", availableGenres[:], &argparse.Options{
	Required: true,
	Help: "Genre to scrape",
})

var concurrency = parser.Int("c", "concurrency", &argparse.Options{
	Help: "Number of connections",
	Default: 4,
})

var minPage = parser.Int("", "min-page", &argparse.Options{
	Help: "Starting page",
	Default: 1,
})

var dir = parser.String("o", "out-dir", &argparse.Options{
	Help: "Output directory",
	Default: "Downloads",
})

var verbose = parser.Flag("v", "verbose", &argparse.Options{
	Help: "More output",
})

func parseArgs() {
	if err := parser.Parse(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*dir, 0777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

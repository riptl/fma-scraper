package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)

var genre = pflag.StringP("genre", "g", "", "Genre to archive")
var listGenres = pflag.Bool("list-genres", false, "Print a list of genres")
var concurrency = pflag.UintP("concurrency", "c", 4, "Number of concurrent downloads")
var minPage = pflag.Uint("min-page", 1, "Starting page")
var pageSize = pflag.Uint("page-size", 500, "Page size")
var dir = pflag.StringP("out-dir", "o", "Downloads", "Output directory")
var verbose = pflag.BoolP("verbose", "v", false, "More output")

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

func parseArgs() error {
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr,
`Free Music Archive Scraper by terorie 2018
 << https://github.com/terorie/fma-scraper >>

Usage:`)
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if *listGenres {
		fmt.Println("Available Genres:")
		for _, g := range availableGenres {
			fmt.Println(g)
		}
		os.Exit(1)
	}

	if *genre == "" {
		return fmt.Errorf("-g/--genre flag required")
	}

	if *concurrency <= 0 {
		return fmt.Errorf("invalid value for --concurrency")
	}

	if err := os.MkdirAll(*dir, 0777); err != nil {
		return err
	}

	return nil
}

package main

import (
	"bufio"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"
)

func logger(results <-chan Track) {
	defer helperGroup.Done()
	f, err := os.OpenFile("downloaded.txt", os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
	if err != nil {
		logrus.Fatal(err)
	}
	defer f.Close()

	wr := bufio.NewWriter(f)
	defer wr.Flush()

	j := json.NewEncoder(wr)

	for result := range results {
		err := j.Encode(result)
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

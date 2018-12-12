package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"mime"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

func downloader(jobs <-chan Track, results chan<- Track) {
	defer downloadGroup.Done()
	for job := range jobs {
		if atomic.LoadInt32(&exitRequested) != 0 {
			break
		}

		now := time.Now()

		u, err := followRedirect(job.Download)
		if err != nil {
			logrus.WithError(err).
				WithField("url", job.Download).
				WithField("title", job.Title).
				Error("Download failed")
		}
		job.Download = u

		n, err := download(u)
		if os.IsExist(err) {
			logrus.
				WithField("title", job.Title).
				Warning("Already downloaded")
		} else if err != nil {
			logrus.
				WithError(err).
				WithField("url", job.Download).
				WithField("title", job.Title).
				Error("Download failed")
		}

		dur := time.Since(now)

		atomic.AddInt64(&totalDownloaded, n)

		logrus.WithFields(logrus.Fields{
			"title": job.Title,
			"size": n,
			"dur": dur.Seconds(),
		}).Info("Downloaded track")

		results <- job
	}
}

func followRedirect(u string) (string, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(u)

	if err := fasthttp.Do(req, res); err != nil {
		return "", err
	}

	if sc := res.StatusCode(); sc != 302 {
		return "", fmt.Errorf("failed to get redirected to mp3: HTTP status %d", sc)
	}

	return string(res.Header.Peek("Location")), nil
}

func download(u string) (int64, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(u)

	if err := fasthttp.Do(req, res); err != nil {
		return 0, err
	}

	if sc := res.StatusCode(); sc != 200 {
		return 0, fmt.Errorf("HTTP status %d", sc)
	}

	cd := string(res.Header.Peek("Content-Disposition"))
	if cd == "" {
		return 0, fmt.Errorf("missing Content-Disposition header")
	}

	_, params, err := mime.ParseMediaType(cd)
	if err != nil {
		return 0, err
	}

	fileName := params["filename"]
	if fileName == "" {
		return 0, fmt.Errorf("missing file name in Content-Disposition header")
	}

	fileName = filepath.Join(*dir, fileName)
	f, err := os.OpenFile(fileName, os.O_CREATE | os.O_EXCL | os.O_WRONLY, 0666)
	if err != nil {
		return 0, err
	}

	return res.WriteTo(f)
}

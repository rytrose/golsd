package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
)

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
// Taken from https://golangcode.com/print-the-current-memory-usage/
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

// Taken from https://golangcode.com/print-the-current-memory-usage/
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Fetches a file from a URL and reads it into memory
func readerFromURL(url string) io.ReadCloser {
	fmt.Println(fmt.Sprintf("Getting: %s", url))
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Sprintf("Couldn't download file %s: %s", url, err))
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Unable to read file to bytes %s: %s", url, err))
	}
	resp.Body.Close()

	return ioutil.NopCloser(bytes.NewReader(fileBytes))
}

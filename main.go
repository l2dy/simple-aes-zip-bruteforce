package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/alexmullins/zip"
)

func worker(content []byte, jobs <-chan int, results chan<- int) {
	size := int64(len(content))
	for i := range jobs {
		password := fmt.Sprintf("%06d", i)
		if testZip(content, size, password) {
			results <- i
		} else {
			results <- -1
		}
	}
}

func testZip(content []byte, size int64, password string) bool {
	r := bytes.NewReader(content)
	reader, err := zip.NewReader(r, size)
	if err != nil {
		log.Print(err)
		return false
	}

	for _, entry := range reader.File {
		entry.SetPassword(password)
		fr, err := entry.Open()
		if err != nil {
			return false
		}
		n, err := io.Copy(io.Discard, fr)
		if n == 0 || err != nil {
			return false
		}
	}
	return true
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <.zip>\n", os.Args[0])
		os.Exit(1)
	}
	zipFile := os.Args[1]
	f, err := os.Open(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	workers := runtime.NumCPU()
	jobs := make(chan int, workers)
	results := make(chan int, workers)

	for w := 0; w < workers; w++ {
		go worker(content, jobs, results)
	}

	const max = 1000000
	recv := 0

	for sent := 0; sent < max; {
		select {
		case jobs <- sent:
			sent++
		case r := <-results:
			recv++
			if r > -1 {
				fmt.Printf("Password is %06d\n", r)
				return
			}
		}
	}

	for ; recv < max; recv++ {
		r := <-results
		if r > -1 {
			fmt.Printf("Password is %06d\n", r)
			return
		}
	}

	fmt.Println("Password not found")
}

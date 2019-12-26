package main

import (
	"fmt"
	"strings"
	"sync"
)

type Reader interface {
	Read(rc chan string)
}

type Writer interface {
	Write(wc chan string)
}

type LogProcess struct {
	rc chan string // channel for read
	wc chan string // channel for write
	reader Reader
	writer Writer
}

type ReadFromFile struct {
	path string
}

type WriteToInflux struct {
	influxDBsn string
}

func (r *ReadFromFile) Read(rc chan string) {
	line := "message"
	rc <- line
}

func (w *WriteToInflux) Write(wc chan string) {
	fmt.Printf("%s\n", <-wc)
}

func (l *LogProcess) Process() {
	// parse the log
	data := <- l.rc
	l.wc <- strings.ToUpper(data)
}



func main() {
	r := &ReadFromFile {
		path: "/tmp/access.log",
	}

	w := &WriteToInflux{
		influxDBsn: "username&password",
	}

	lp := &LogProcess {
		rc:	make(chan string),
		wc: make(chan string),
		reader: r,
		writer: w,
	}


	// concurrently handle three functionality
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		lp.reader.Read(lp.rc)
		wg.Done()
	}()
	go func() {
		lp.Process()
		wg.Done()
	}()
	go func() {
		lp.writer.Write(lp.wc)
		wg.Done()
	}()

	wg.Wait()
}

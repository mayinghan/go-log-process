package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Reader interface {
	Read(rc chan []byte)
}

type Writer interface {
	Write(wc chan string)
}

type LogProcess struct {
	rc chan []byte // channel for read
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

func (r *ReadFromFile) Read(rc chan []byte) {
	f, err := os.Open(r.path)

	if err != nil {
		panic(errors.New("open file error"))
	}

	// read from the end of the file to make this a real-time monitor
	f.Seek(0, 2) // move the read pointer to the  end of the file
	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(5000) // wait for new logs
			continue
		} else if err != nil {
			panic(errors.New("error happened when reading by line"))
		}
		rc <- line[:len(line) - 1]
	}

}

func (w *WriteToInflux) Write(wc chan string) {
	for v := range wc {
		fmt.Printf("%s\n", v)
	}
}

func (l *LogProcess) Process() {
	// parse the log

	for v := range l.rc {
		l.wc <- strings.ToUpper(string(v))
	}

}


func main() {
	r := &ReadFromFile {
		path: "fakedata/access.log",
	}

	w := &WriteToInflux{
		influxDBsn: "username&password",
	}

	lp := &LogProcess {
		rc:	make(chan []byte),
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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)
type Mesasge struct {
	TimeLocal time.Time
	BytesSent int
	Path, Method, Scheme, Status string
	UpstreamTime, RequestTime float64
}

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

func (l *LogProcess) Process() {
	// parse the log
	/**
	172.0.0.12 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	*/
	loc, _ := time.LoadLocation("America/New_York")
	re := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)
	for v := range l.rc {
		matchedGroup := re.FindStringSubmatch(string(v))
		if len(matchedGroup) != 13 {
			log.Println("Match failed", string(v))
			continue
		}

		msg := &Mesasge{}
		t, err := time.ParseInLocation("02/Jan/2006: 15:04:05 +0000", matchedGroup[4], loc)
		if err != nil {
			log.Println("ParseInLocation failed", err.Error(), matchedGroup[4])
		}

		msg.TimeLocal = t
		byteSent, _ := strconv.Atoi(matchedGroup[8])
		msg.BytesSent = byteSent
		l.wc <- strings.ToUpper(string(v))
	}

}

func (w *WriteToInflux) Write(wc chan string) {
	for v := range wc {
		fmt.Printf("%s\n", v)
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

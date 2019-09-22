bcast
=====

[![GoDoc](http://godoc.org/github.com/DeedleFake/bcast?status.svg)](http://godoc.org/github.com/DeedleFake/bcast)
[![Go Report Card](https://goreportcard.com/badge/github.com/DeedleFake/bcast)](https://goreportcard.com/report/github.com/DeedleFake/bcast)

bcast is a simple N:M channel-based broadcasting system for Go. It attempts to be a generic stand-in for the lack of such capability with just plain channels.

Example
-------

```go
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/DeedleFake/bcast"
)

var bc bcast.Broadcast

func client(c net.Conn) {
	defer c.Close()

	type msg struct {
		from net.Addr
		text string
	}

	in := make(chan interface{})
	stop := bc.Listen(in)
	defer stop()

	go func() {
		for data := range in {
			msg := data.(*msg)
			if msg.from == c.RemoteAddr() {
				continue
			}

			_, err := io.WriteString(c, msg.text)
			if err != nil {
				log.Printf("Client write error: %v", err)
				return
			}
		}
	}()

	r := bufio.NewReader(c)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Printf("Client read error: %v", err)
			return
		}
		if len(line) == 1 {
			continue
		}

		bc.Send() <- &msg{
			from: c.RemoteAddr(),
			text: fmt.Sprintf("%v> %v", c.RemoteAddr(), line),
		}
	}

	log.Printf("Client disconnected: %v", c.RemoteAddr())
}

func main() {
	addr := flag.String("addr", ":12345", "address to listen on")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Failed to open listener: %v", err)
	}
	defer lis.Close()

	for {
		c, err := lis.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Client connected: %v", c.RemoteAddr())

		go client(c)
	}
}
```

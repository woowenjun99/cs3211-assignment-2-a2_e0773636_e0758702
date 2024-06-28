package main

import "C"
import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const DEBUGPRINT = false

type Engine struct {
	done        chan interface{}
	inputStream chan InputOrder
}

func (e *Engine) init() {
	e.done = make(chan interface{})
	e.inputStream = make(chan InputOrder)
	go instructionTask(e.done, e.inputStream)
}

func (e *Engine) end() {
	<-e.done
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(conn, e)
}

func handleConn(conn net.Conn, e *Engine) {
	defer conn.Close()
	for {
		in, err := readInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		// put into inputstream
		// fmt.Fprintf(os.Stderr, "	Sending order: %c %v x %v @ %v ID: %v\n", in.orderType, in.instrument, in.count, in.price, in.orderId)
		i := InputOrder{
			order:     in,
			timestamp: GetCurrentTimestamp(),
			wait:      make(chan bool, 1),
		}

		e.inputStream <- i
		<-i.wait
	}
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

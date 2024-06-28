package main

import (
	"fmt"
	"os"
)

type InputOrder struct {
	order     input
	timestamp int64
	wait      chan bool
}

// exchange receive incoming order
// routes to respective instrument's command channel
func instructionTask(done <-chan interface{}, inputStream chan InputOrder) {
	instrMap := make(map[string]chan InputOrder)
	orderCache := make(map[uint32]string)
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "	Initialised instruction task\n")
	}
	for {
		select {
		case i := <-inputStream:
			cmd := i.order
			var instrChan chan InputOrder
			if cmd.orderType == 'C' {
				instr, exists := orderCache[cmd.orderId]
				if !exists {
					// reject if order does not exist
					outputOrderDeleted(
						cmd,
						false,
						i.timestamp,
					)
				}
				instrChan = instrMap[instr]
			} else {
				// route to instrument
				channel, exists := instrMap[cmd.instrument]
				if !exists {
					instrChan = make(chan InputOrder)

					go orderTask(done, instrChan)
					instrMap[cmd.instrument] = instrChan
				} else {
					instrChan = channel
				}

				orderCache[cmd.orderId] = cmd.instrument
			}
			instrChan <- i

		case <-done:
			return
		}
	}
}

// instrument listens to incoming order from command channel
// executes match or cancel
func orderTask(done <-chan interface{}, inputStream chan InputOrder) {
	// listens to incoming order, executes the order
	orderBook := InitOrders()
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "	Initialised order task\n")
	}

	for {
		select {
		case inp := <-inputStream:
			cmd := inp.order
			if DEBUGPRINT {
				fmt.Fprintf(os.Stderr, "orderTask order: %c %v x %v @ %v ID: %v\n", cmd.orderType, cmd.instrument, cmd.count, cmd.price, cmd.orderId)
			}
			switch cmd.orderType {
			case 'B':
				// handle buy
				MatchBOrder(&orderBook, cmd, inp.timestamp)

			case 'S':
				// handle sell
				MatchSOrder(&orderBook, cmd, inp.timestamp)

			case 'C':
				// handle cancel
				CancelSBOrder(&orderBook, cmd, inp.timestamp)

			}
			inp.wait <- true
		case <-done:
			return
		}
	}
}

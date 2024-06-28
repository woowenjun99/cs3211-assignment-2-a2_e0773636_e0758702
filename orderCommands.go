package main

import (
	"container/heap"
	"fmt"
	"os"
)

func CancelSBOrder(o *Orders, cmd input, timestamp int64) {
	// search thru the heap to find id
	// remove that id
	tmpQueue := []*Node{}
	found := false
	isAccepted := true
	id := cmd.orderId

	isSell, exists := o.sbCache[cmd.orderId]
	if exists {
		if isSell {
			// SELL SIDE
			pq := o.sellHeap
			for _, node := range *pq {
				if node.orderId == id {
					found = true
					break
				}
			}
			if !found {
				isAccepted = false
			} else {
				for pq.Len() > 0 {
					node := heap.Pop(pq).(*Node)
					if node.orderId == id {
						continue
					} else {
						tmpQueue = append(tmpQueue, node)
					}
				}

				for _, node := range tmpQueue {
					heap.Push(pq, node)
				}
			}
			if DEBUGPRINT {
				fmt.Fprintf(os.Stderr, "		pq len: %v\n", pq.Len())
			}
		} else {
			// BUY SIDE
			pq := o.buyHeap
			for _, node := range *pq {
				if node.orderId == id {
					found = true
					break
				}
			}
			if !found {
				isAccepted = false
			} else {
				for pq.Len() > 0 {
					node := heap.Pop(pq).(*Node)
					if node.orderId == id {
						continue
					} else {
						tmpQueue = append(tmpQueue, node)
					}
				}

				for _, node := range tmpQueue {
					heap.Push(pq, node)
				}
			}
			if DEBUGPRINT {
				fmt.Fprintf(os.Stderr, "		pq len: %v\n", pq.Len())
			}
		}

	}

	outputOrderDeleted(
		cmd,
		isAccepted,
		timestamp,
	)
}

func MatchSOrder(o *Orders, cmd input, timestamp int64) {
	is_sell := cmd.orderType == 'S'
	tmpQueue := []*Node{}
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		MatchOrder matching: %v %c price %v count %v\n", cmd.orderId, cmd.orderType, cmd.price, cmd.count)
	}
	pq := o.buyHeap
	o.sbCache[cmd.orderId] = true
	// if sell, take the price of a buy node at price >= cmd.price
	// keep popping until reach the price
	// if buy, take the lowest price of a sell node up to the cmd.price
	// take the first price, if not keep popping
	for pq.Len() > 0 {
		node := heap.Pop(pq).(*Node)
		condition := node.price >= cmd.price
		if !is_sell {
			condition = node.price <= cmd.price
		}
		if DEBUGPRINT {
			fmt.Fprintf(os.Stderr, "		currentnode: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
		}
		if condition && node.isSellSide == !is_sell {
			// check count
			if node.count > cmd.count {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		full match: node count minus\n")
				}
				// if node > cmd, decrement node and exit
				// 1 execute only
				node.count -= cmd.count
				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					cmd.count,
					timestamp,
				)
				node.executionId++
				cmd.count = 0
				// add this node back but stop finding match
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		adding node to tmpqueue: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
				}
				tmpQueue = append(tmpQueue, node)
				break
			} else if node.count < cmd.count {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		partial match\n")
				}
				// if cmd > node count, decrement cmd delete node and continue
				// multiple executes possible
				cmd.count -= node.count
				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					node.count,
					timestamp,
				)
				node.executionId++
				node.count = 0
				// continue finding match, but dont add this node back
			} else {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		full match: node delete\n")
				}
				// if node = cmd, delete node and input, exit
				// 1 execute only
				node.count = 0
				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					cmd.count,
					timestamp,
				)
				cmd.count = 0
				break
				// both should be deleted and stop finding match
			}
		}
		if node.count > 0 {
			if DEBUGPRINT {
				fmt.Fprintf(os.Stderr, "		adding node to tmpqueue: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
			}
			tmpQueue = append(tmpQueue, node)
		}
	}
	// breaks out here, add node if not fully matched and put back other nodes into pq
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		exit loop tmparray %v\n", len(tmpQueue))
	}
	// add as a node if cmd still leftover
	if cmd.count > 0 {
		if DEBUGPRINT {
			fmt.Fprintf(os.Stderr, "		add new node: ID %v x %v\n", cmd.orderId, cmd.count)
		}
		newNode := Node{
			orderId:     cmd.orderId,
			price:       cmd.price,
			count:       cmd.count,
			timestamp:   timestamp,
			isSellSide:  is_sell,
			executionId: 1,
		}

		heap.Push(o.sellHeap, &newNode)
		outputOrderAdded(cmd, timestamp)
	}

	for _, node := range tmpQueue {
		heap.Push(pq, node)
	}
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		pq len: %v\n", pq.Len())
	}
}

func MatchBOrder(o *Orders, cmd input, timestamp int64) {
	is_sell := cmd.orderType == 'S'
	tmpQueue := []*Node{}
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		MatchOrder matching: %v %c price %v count %v\n", cmd.orderId, cmd.orderType, cmd.price, cmd.count)
	}

	pq := o.sellHeap
	o.sbCache[cmd.orderId] = false

	// if sell, take the price of a buy node at price >= cmd.price
	// keep popping until reach the price
	// if buy, take the lowest price of a sell node up to the cmd.price
	// take the first price, if not keep popping
	for pq.Len() > 0 {
		node := heap.Pop(pq).(*Node)
		condition := node.price >= cmd.price
		if !is_sell {
			condition = node.price <= cmd.price
		}
		if DEBUGPRINT {
			fmt.Fprintf(os.Stderr, "		currentnode: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
		}
		if condition && node.isSellSide == !is_sell {
			// check count
			if node.count > cmd.count {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		full match: node count minus\n")
				}
				// if node > cmd, decrement node and exit
				// 1 execute only
				node.count -= cmd.count
				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					cmd.count,
					timestamp,
				)
				node.executionId++
				cmd.count = 0
				// add this node back but stop finding match
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		adding node to tmpqueue: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
				}
				tmpQueue = append(tmpQueue, node)
				break
			} else if node.count < cmd.count {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		partial match\n")
				}
				// if cmd > node count, decrement cmd delete node and continue
				// multiple executes possible
				cmd.count -= node.count

				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					node.count,
					timestamp,
				)
				node.executionId++
				node.count = 0
				// continue finding match, but dont add this node back
			} else {
				if DEBUGPRINT {
					fmt.Fprintf(os.Stderr, "		full match: node delete\n")
				}
				// if node = cmd, delete node and input, exit
				// 1 execute only
				node.count = 0

				outputOrderExecuted(
					node.orderId,
					cmd.orderId,
					node.executionId,
					node.price,
					cmd.count,
					timestamp,
				)
				cmd.count = 0
				break
				// both should be deleted and stop finding match
			}
		}
		if node.count > 0 {
			if DEBUGPRINT {
				fmt.Fprintf(os.Stderr, "		adding node to tmpqueue: sell? %v Count: %v Price: %v ID: %v\n", node.isSellSide, node.count, node.price, node.orderId)
			}
			tmpQueue = append(tmpQueue, node)
		}
	}
	// breaks out here, add node if not fully matched and put back other nodes into pq
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		exit loop tmparray %v\n", len(tmpQueue))
	}
	// add as a node if cmd still leftover
	if cmd.count > 0 {
		if DEBUGPRINT {
			fmt.Fprintf(os.Stderr, "		add new node: ID %v x %v\n", cmd.orderId, cmd.count)
		}
		newNode := Node{
			orderId:     cmd.orderId,
			price:       cmd.price,
			count:       cmd.count,
			timestamp:   timestamp,
			isSellSide:  is_sell,
			executionId: 1,
		}
		heap.Push(o.buyHeap, &newNode)
		outputOrderAdded(cmd, timestamp)
	}

	for _, node := range tmpQueue {
		heap.Push(pq, node)
	}
	if DEBUGPRINT {
		fmt.Fprintf(os.Stderr, "		pq len: %v\n", pq.Len())
	}
}

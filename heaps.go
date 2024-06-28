package main

import (
	"container/heap"
)

type Node struct {
	orderId     uint32
	price       uint32
	count       uint32
	timestamp   int64
	isSellSide  bool
	executionId uint32
}

/**
Regular Heap
*/

type OrderHeap []*Node

func (pq OrderHeap) Len() int { return len(pq) }
func (pq OrderHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq OrderHeap) Less(i, j int) bool {
	// If prices are the same, sort by timestamp in ascending order
	if pq[i].price == pq[j].price {
		return pq[i].timestamp < pq[j].timestamp
	} else {
		// Sort by price in descending order
		return pq[i].price < pq[j].price
	}
}
func (pq *OrderHeap) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
func (pq *OrderHeap) Push(x any) {
	item := x.(*Node)
	*pq = append(*pq, item)
}

/**
Buy and Sell Heaps in different ordering
*/

type BuyHeap []*Node
type SellHeap []*Node
type Orders struct {
	buyHeap  *BuyHeap
	sellHeap *SellHeap
	sbCache  map[uint32]bool
}

func InitOrders() Orders {
	buys := make(BuyHeap, 0)
	sells := make(SellHeap, 0)
	heap.Init(&buys)
	heap.Init(&sells)

	cache := make(map[uint32]bool)

	orders := Orders{
		buyHeap:  &buys,
		sellHeap: &sells,
		sbCache:  cache,
	}
	return orders
}

func (pq BuyHeap) Len() int { return len(pq) }
func (pq BuyHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq BuyHeap) Less(i, j int) bool {
	// If prices are the same, sort by timestamp in ascending order
	if pq[i].price == pq[j].price {
		return pq[i].timestamp < pq[j].timestamp
	} else {
		// Sort by price in descending order
		return pq[i].price > pq[j].price
	}
}
func (pq *BuyHeap) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
func (pq *BuyHeap) Push(x any) {
	item := x.(*Node)
	*pq = append(*pq, item)
}

func (pq SellHeap) Len() int { return len(pq) }
func (pq SellHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq SellHeap) Less(i, j int) bool {
	// If prices are the same, sort by timestamp in ascending order
	if pq[i].price == pq[j].price {
		return pq[i].timestamp < pq[j].timestamp
	} else {
		// Sort by price in ascending order
		return pq[i].price < pq[j].price
	}
}
func (pq *SellHeap) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
func (pq *SellHeap) Push(x any) {
	item := x.(*Node)
	*pq = append(*pq, item)
}

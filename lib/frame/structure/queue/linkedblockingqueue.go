package queue

import "math"

type node struct {
	item interface{}
	next *node
}

func makeNode(item interface{}) *node {
	return &node{
		item: item,
	}
}

type LinkedBlockingQueue struct {
	// 容量
	capacity int64

	// 总数
	count int64

	head *node

	last *node

	putlock chan int

	takelock chan int
}

func MakeLinkedBlockingQueue(capacity int64) *LinkedBlockingQueue {
	if capacity == 0 {
		capacity = math.MaxInt64
	}
	return &LinkedBlockingQueue{
		capacity: capacity,
		count:    0,
		head:     nil,
		last:     nil,
		putlock:  make(chan int),
		takelock: make(chan int),
	}
}

// Put 队列尾部插入元素
func (l *LinkedBlockingQueue) Put(item interface{}) {

}

// Take 队列头部取出数据
func (l *LinkedBlockingQueue) Take(seconds int) interface{} {
	return nil
}

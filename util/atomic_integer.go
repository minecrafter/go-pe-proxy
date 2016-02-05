package util

import "sync/atomic"

type AtomicInteger struct {
	val int32
}

func (ai *AtomicInteger) IncrementAndGet() int32 {
	return atomic.AddInt32(&ai.val, 1)
}

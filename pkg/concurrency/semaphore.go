package concurrency

import "math"

const (
	DefaultConcurrency = 500
	MaxConcurrency     = math.MaxUint16
)

//Semaphore ctl the max concurrency.
type Semaphore struct {
	tickets chan bool
}

//NewSemaphore accept concurrency number, not more than 65535
func NewSemaphore(concurrency int) *Semaphore {
	if concurrency >= math.MaxUint16 {
		concurrency = MaxConcurrency
	}
	b := &Semaphore{
		tickets: make(chan bool, concurrency),
	}
	for i := 0; i < concurrency; i++ {
		b.tickets <- true
	}
	return b
}
func (b *Semaphore) Acquire() {
	<-b.tickets
}

//Release return back signal.
func (b *Semaphore) Release() {
	b.tickets <- true
}

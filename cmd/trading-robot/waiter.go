package main

import "sync"

type Waiter struct {
	wg sync.WaitGroup
}

func (w *Waiter) Go(f func()) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		f()
	}()
}

func (w *Waiter) Wait() {
	w.wg.Wait()
}

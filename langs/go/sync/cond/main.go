package main

import (
	"fmt"
	"sync"
	"time"
)

var sharedRsc = make(map[string]interface{})

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	m := sync.Mutex{}
	c := sync.NewCond(&m)

	go func() {
		// This goroutine waits for changes to sharedRsc
		c.L.Lock()
		for len(sharedRsc) == 0 || sharedRsc["rsc1"] == nil {
			c.Wait()
		}
		fmt.Println("goroutine1:", sharedRsc["rsc1"])

		go func() {
			fmt.Println("goroutine1: sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			c.Broadcast() // Signal goroutine2 to proceed
		}()
		c.L.Unlock()
		wg.Done()
	}()

	go func() {
		// This goroutine waits until goroutine1 finishes
		c.L.Lock()
		for !c.Signal() {
			fmt.Println("can't proceed yet (goroutine2)")
			c.Wait()
		}
		fmt.Println("goroutine2:", sharedRsc["rsc2"])
		c.L.Unlock()
		wg.Done()
	}()

	// This part writes changes to sharedRsc
	c.L.Lock()
	sharedRsc["rsc1"] = "foo"
	sharedRsc["rsc2"] = "bar"
	c.Broadcast() // Signal goroutine1 to proceed
	c.L.Unlock()

	wg.Wait()
}

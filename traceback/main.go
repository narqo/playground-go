package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	go func() {
		for {
			log.Printf("level=info running")
			time.Sleep(time.Second / 3)
		}
	}()

	time.Sleep(time.Second)

	runmain(2)
}

var race bool

func runmain(n int) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func() {
			defer wg.Done()

			race = true

			for {
				log.Printf("level=info worker %d running", i)
				time.Sleep(time.Second / 3)
				if i > 0 && rand.IntN(100) > 60 {
					panic(fmt.Sprintf("worker %d: something went\n very\n wrong", i))
				}
			}
		}()
	}
	wg.Wait()
}

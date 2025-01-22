package circuitbreaker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

func SimulateCircuitBreaker() {
	count := 1

	args := os.Args[1:]
	if len(args) != 0 {
		i, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("The argument must be an integer")
			return
		}
		count = i
	}

	cb := NewCircuitBreaker(3, 200*time.Millisecond)
	wg := sync.WaitGroup{}

	for i := 1; i <= count; i++ {
		wg.Add(1)
		go callDownstream(i, cb, &wg)
	}

	wg.Wait()
}

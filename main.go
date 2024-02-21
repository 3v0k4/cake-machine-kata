package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func orDone(done, channel <-chan interface{}) <-chan interface{} {
	stream := make(chan interface{})

	go func() {
		defer close(stream)

		for {
			select {
			case <-done:
				return
			case value, ok := <-channel:
				if !ok {
					return
				}
				select {
				case <-done:
				case stream <- value:
				}
			}
		}
	}()

	return stream
}

func fanIn(done <-chan interface{}, channels []<-chan interface{}) <-chan interface{} {
	stream := make(chan interface{})

	var wg sync.WaitGroup

	multiplex := func(channel <-chan interface{}) {
		defer wg.Done()
		for value := range orDone(done, channel) {
			select {
			case <-done:
				return
			case stream <- value:
			}
		}
	}

	for _, channel := range channels {
		wg.Add(1)
		go multiplex(channel)
	}

	go func() {
		defer close(stream)
		wg.Wait()
	}()

	return stream
}

func tee(done, channel <-chan interface{}) (_, _ <-chan interface{}) {
	stream1 := make(chan interface{})
	stream2 := make(chan interface{})

	go func() {
		defer close(stream1)
		defer close(stream2)

		for value := range orDone(done, channel) {
			var stream1 = stream1
			var stream2 = stream2

			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case stream1 <- value:
					stream1 = nil
				case stream2 <- value:
					stream2 = nil
				}
			}
		}
	}()

	return stream1, stream2
}

func prep(done <-chan interface{}) <-chan interface{} {
	duration := time.Duration(5+rand.Intn(4)) * time.Second
	stream := make(chan interface{})

	go func() {
		defer close(stream)

		for {
			select {
			case <-done:
				return
			case <-time.After(duration):
				select {
				case <-done:
					return
				case stream <- struct{}{}:
				}
			}
		}
	}()

	return stream
}

func cook(done, start <-chan interface{}) <-chan interface{} {
	duration := 10 * time.Second
	stream := make(chan interface{})

	go func() {
		defer close(stream)

		for range orDone(done, start) {
			select {
			case <-done:
				return
			case <-time.After(duration):
				select {
				case <-done:
					return
				case stream <- struct{}{}:
				}
			}
		}
	}()

	return stream
}

func pack(done, start <-chan interface{}) <-chan interface{} {
	duration := 2 * time.Second
	stream := make(chan interface{})

	go func() {
		defer close(stream)

		for range orDone(done, start) {
			select {
			case <-done:
				return
			case <-time.After(duration):
				select {
				case <-done:
					return
				case stream <- struct{}{}:
				}
			}
		}
	}()

	return stream
}

func serial() {
	done := make(chan interface{})
	go func() {
		<-time.After(5*time.Minute + 10*time.Second)
		close(done)
	}()

	prep := prep(done)
	prep1, prep2 := tee(done, prep)

	cook := cook(done, prep1)
	cook1, cook2 := tee(done, cook)

	pack := pack(done, cook1)
	pack1, pack2 := tee(done, pack)

	go func() {
		var prepped, cooked, packed int
		tick := time.Tick(1 * time.Minute)

		for {
			select {
			case <-done:
				return
			case <-prep2:
				prepped++
			case <-cook2:
				cooked++
			case <-pack2:
				packed++
			case <-tick:
				fmt.Printf("prepped: %d, cooked: %d, packed: %d\n", prepped-cooked, cooked-packed, packed)
			}
		}
	}()

	now := time.Now()
	for range pack1 {
	}
	fmt.Println(time.Since(now))
}

func parallel() {
	done := make(chan interface{})
	go func() {
		<-time.After(5*time.Minute + 10*time.Second)
		close(done)
	}()

	preps := make([]<-chan interface{}, 3)
	for i := 0; i < 3; i++ {
		preps[i] = prep(done)
	}
	prep := fanIn(done, preps)
	prep1, prep2 := tee(done, prep)

	cooks := make([]<-chan interface{}, 5)
	for i := 0; i < 5; i++ {
		cooks[i] = cook(done, prep1)
	}
	cook := fanIn(done, cooks)
	cook1, cook2 := tee(done, cook)

	packs := make([]<-chan interface{}, 2)
	for i := 0; i < 2; i++ {
		packs[i] = pack(done, cook1)
	}
	pack := fanIn(done, packs)
	pack1, pack2 := tee(done, pack)

	go func() {
		var prepped, cooked, packed int
		tick := time.Tick(1 * time.Minute)

		for {
			select {
			case <-done:
				return
			case <-prep2:
				prepped++
			case <-cook2:
				cooked++
			case <-pack2:
				packed++
			case <-tick:
				fmt.Printf("prepped: %d, cooked: %d, packed: %d\n", prepped-cooked, cooked-packed, packed)
			}
		}
	}()

	now := time.Now()
	for range pack1 {
	}
	fmt.Println(time.Since(now))
}

func main() {
	serial()
	parallel()
}

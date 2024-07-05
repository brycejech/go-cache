package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/brycejech/go-cache/cache"
	"golang.org/x/exp/rand"
)

func main() {
	cache := cache.NewCache()

	keys := []string{
		"Key1",
		"Key2",
		"Key3",
		"Key4",
		"Key5",
		"Key6",
		"Key7",
	}
	messages := []string{
		"Message 1",
		"Message 2",
		"Message 3",
		"Message 4",
		"Message 5",
	}

	rand.Seed(uint64(time.Now().Unix()))

	var (
		writeDur   time.Duration
		readDur    time.Duration
		wg         = &sync.WaitGroup{}
		iterations = 1_000_00
	)

	writeCh := make(chan time.Duration, iterations)
	readCh := make(chan time.Duration, iterations)

	writeTest(keys, messages, iterations, cache, writeCh, wg)
	readTest(keys, iterations, cache, readCh, wg)

	go func() {
		wg.Wait()
		close(writeCh)
		close(readCh)
	}()

	for dur := range writeCh {
		writeDur += dur
	}
	for dur := range readCh {
		readDur += dur
	}

	fmt.Println("writeDurAvg:", float64((writeDur/time.Millisecond))/float64(iterations), "ms")
	fmt.Println("readDurAvg:", float64((readDur/time.Millisecond))/float64(iterations), "ms")
}

func writeTest(
	keys []string,
	messages []string,
	iterations int,
	cache cache.Cache,
	ch chan time.Duration,
	wg *sync.WaitGroup,
) {
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path := getRandPath(keys)
			message := pickRandom(messages)

			start := time.Now()
			cache.Set(path, 100, []byte(message))
			dur := time.Since(start)
			ch <- dur
		}()
	}
}

func readTest(
	keys []string,
	iterations int,
	cache cache.Cache,
	ch chan time.Duration,
	wg *sync.WaitGroup,
) {
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path := getRandPath(keys)

			start := time.Now()
			item := cache.Get(path)
			if item != nil {
				item.Read()
				dur := time.Since(start)
				ch <- dur
			} else {
				dur := time.Since(start)
				ch <- dur
			}
		}()
	}
}

func getRandPath(keys []string) []string {
	pathLen := rand.Intn(len(keys)) + 1
	path := make([]string, pathLen)

	for ii := 0; ii < pathLen; ii++ {
		path[ii] = pickRandom(keys)
	}

	return path
}

func pickRandom[T any](items []T) T {
	idx := rand.Intn(len(items))

	return items[idx]
}

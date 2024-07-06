package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/brycejech/go-cache/cache"
	"github.com/brycejech/go-cache/util"
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
		"Sustainable shabby chic deep v, banjo tote bag vibecession direct trade humblebrag viral JOMO 90's affogato gentrify food truck hoodie",
		"Bodega boys affogato poutine beard la croix kombucha butcher",
		"Squid gluten-free echo park chillwave locavore authentic hexagon tacos quinoa big mood plaid taxidermy kitsch bitters fit",
		"Marxism everyday carry wolf, brunch direct trade fingerstache ascot snackwave palo santo 90's taxidermy gatekeep synth distillery",
		"Mustache neutral milk hotel squid yuccie forage coloring book, four dollar toast kogi try-hard adaptogen",
		"iPhone occupy kogi food truck, brunch la croix viral post-ironic venmo fam irony artisan",
		"Pok pok kickstarter chambray, vexillologist franzen semiotics paleo tacos",
	}

	rand.Seed(uint64(time.Now().Unix()))

	var (
		writeDur   time.Duration
		readDur    time.Duration
		wg         = &sync.WaitGroup{}
		iterations = 1_000_000
	)

	writeCh := make(chan time.Duration, iterations)
	readCh := make(chan time.Duration, iterations)

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan int)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("cacheSize:", util.ByteSizeToStr(int64(cache.Size())))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	writeTest(keys, messages, iterations, cache, writeCh, wg)
	readTest(keys, iterations, cache, readCh, wg)

	func() {
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

	quit <- 0

	fmt.Println("cacheSize:", util.ByteSizeToStr(int64(cache.Size())))
	fmt.Println("writeDurAvg:", float64((writeDur/time.Millisecond))/float64(iterations), "ms")
	fmt.Println("readDurAvg:", float64((readDur/time.Millisecond))/float64(iterations), "ms")
	treeBytes, _ := json.MarshalIndent(cache.Visualize(), "", "  ")
	fmt.Println("Tree:")
	fmt.Println(string(treeBytes))

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
			cache.Set(path, 10_000 /* 10s */, []byte(message))
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
				item.Read() // account for Read() time in real scenario
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

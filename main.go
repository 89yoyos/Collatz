package Collatz

import (
	"bufio"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
	"os"
	"sync"
	"time"
)

type collatzCalculator struct {
	mask        uint64
	maskPower   uint64
	maskIsSolid bool

	Printer chan string
	results chan uint64
}

func NewCollatzCalculator() *collatzCalculator {
	c := collatzCalculator{}
	c.Reset()

	c.Printer = make(chan string)
	c.results = make(chan uint64)
	go c.ResultsWatcher()
	go c.PrinterWatcher()
	c.results <- 2
	return &c
}

func (c collatzCalculator) Reset() {
	c.mask = 2
	c.maskPower = 2
	c.maskIsSolid = false
}

func (c collatzCalculator) BenchmarkSequential(toPower uint64) {
	p := message.NewPrinter(language.English)

	max := p.Sprintf("%d", uint64(2)<<(toPower-1))
	fmt.Printf("Sequentially Calculating Collatz to %s (2^%d)...\n\n", max, toPower)
	start := time.Now()
	c.TestSequentially(toPower)
	done := p.Sprintf("%s\n", time.Since(start))
	fmt.Printf("Sequentially Calculated Collatz to %s (2^%d) in %s\n\n", max, toPower, done)
}

func (c collatzCalculator) BenchmarkConcurrent(toPower uint64) {
	p := message.NewPrinter(language.English)
	max := p.Sprintf("%d", uint64(2)<<(toPower-1))
	fmt.Printf("Concurrently Calculating Collatz to %s (2^%d)...\n\n", max, toPower)
	start := time.Now()
	c.TestConcurrently(toPower)
	done := p.Sprintf("%s\n", time.Since(start))
	fmt.Printf("Concurrently Calculated Collatz to %s (2^%d) in %s\n\n", max, toPower, done)
}

func (c collatzCalculator) ResultsWatcher() {
	if c.maskPower == 0 {
		c.maskPower = 2
	}
	msg := uint64(1)
	for msg != 0 {
		msg = <-c.results
		c.mask = c.mask | msg
		for c.maskPower < msg-1 {
			c.maskPower = c.maskPower << 1
		}
		if c.mask+2 == c.maskPower<<1 {
			c.maskIsSolid = true
		}
		// fmt.Printf("Finished: %d\n",msg)
		// fmt.Printf("Mask:  %s (%d)\n", strconv.FormatUint(c.mask, 2), c.mask)
		// fmt.Printf("Power: %s (%d)\n", strconv.FormatUint(c.maskPower, 2), c.maskPower)
	}
}

func (c collatzCalculator) PrinterWatcher() {
	f, _ := os.Create("c://dat.txt")
	defer f.Close()
	msg := "a"
	w := bufio.NewWriter(f)
	for msg != "stop" {
		msg = <-c.Printer
		_, _ = w.WriteString(msg)
		// _ = w.Flush()
	}
	_ = w.Flush()
}

func (c collatzCalculator) hasProven(n uint64) bool {
	if n > c.mask {
		return false
	}
	if c.maskIsSolid && n < c.mask {
		return true
	}
	x := c.maskPower
	for x > 1 {
		if (x&c.mask)&n != 0 {
			return true
		}
		x = x >> 1
	}
	return false
}

func (c collatzCalculator) TestSequentially(toPower uint64) {
	for i := uint64(2); i < toPower; i++ {

		start := (uint64(2) << (i - 2)) + 1
		stop := uint64(2) << (i - 1)
		for ii := start; ii < stop; ii += 2 {
			c.Test(ii)
		}
		c.results <- stop
	}
}

func (c collatzCalculator) TestConcurrently(toPower uint64) bool {
	var wg sync.WaitGroup
	for i := uint64(2); i < toPower; i++ {
		wg.Add(1)
		start := (uint64(2) << (i - 2)) + 1
		stop := uint64(2) << (i - 1)
		go c.chunkSplitter(start, stop, &wg)
	}
	wg.Wait()
	return true
}

func (c collatzCalculator) chunkSplitter(start uint64, stop uint64, wg *sync.WaitGroup) {
	start = start | 1
	difference := stop - start
	maxChunkSize := uint64(math.Max(float64(difference/14), float64(10000)))
	cursor := start
	var wg2 sync.WaitGroup
	for cursor < stop {
		wg2.Add(1)
		cut := cursor + maxChunkSize - 1
		if cursor+maxChunkSize > stop {
			cut = stop
		}
		go c.calculateChunk(cursor, cut, &wg2)
		cursor += maxChunkSize
	}
	wg2.Wait()
	wg.Done()
	c.results <- stop
}

func (c collatzCalculator) calculateChunk(start uint64, stop uint64, wg *sync.WaitGroup) {
	for i := start; i < stop; i += 2 {
		c.Test(i)
	}
	wg.Done()
}

func (c collatzCalculator) Test(number uint64) {
	// start:=number
	c.Printer <- fmt.Sprintf("%d", number)
	for number > 1 {
		number = c.GetNextCollatzNumber(number)
		c.Printer <- fmt.Sprintf(",%d", number)
	}
	c.Printer <- fmt.Sprintf(",\n")
}

func (c collatzCalculator) GetSteps(number uint64) (steps uint64) {
	steps = 0
	for number > 1 {
		steps++
		number = c.GetNextCollatzNumber(number)
	}
	return steps
}

func (c collatzCalculator) GetNextCollatzNumber(number uint64) uint64 {
	if number%2 == 0 {
		return number >> 1
	}
	return (number << 1) - (number >> 1)
}

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

/*

This program was written by Alyx Green in November 2020
Written for Evan Goff's Paper on the Collatz Conjecture
Designed to be Self-Documenting, but an example of use:
func main() {
	c := Collatz.NewCollatzCalculator()
	c.BenchmarkSequential(24)
}
Definitely not the most efficient possible way, but far
superior to his Python implementation. Cut >13m Python
implementation to ~7ms here on same numbers and PC.

*/

type collatzCalculator struct {
	PrintResults bool
	Printer      chan string
}

func NewCollatzCalculator() *collatzCalculator {
	c := collatzCalculator{}
	c.PrintResults = true
	c.Printer = make(chan string)
	go c.PrinterWatcher()
	return &c
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

func (c collatzCalculator) PrinterWatcher() { // things to write to file
	f, _ := os.Create("c://dat.txt") // replace as needed
	defer f.Close()
	msg := "a" // initialise
	w := bufio.NewWriter(f)
	for msg != "stop" {
		msg = <-c.Printer
		_, _ = w.WriteString(msg)
		_ = w.Flush()
	}
	_ = w.Flush()
}

func (c collatzCalculator) TestSequentially(toPower uint64) { // no concurrency
	for i := uint64(2); i < toPower; i++ {
		// splits by power of 2 (remnant of old method of checking progress)
		start := (uint64(2) << (i - 2)) + 1
		stop := uint64(2) << (i - 1)
		for ii := start; ii < stop; ii += 2 {
			c.Test(ii)
		}
	}
}

func (c collatzCalculator) TestConcurrently(toPower uint64) bool { // with concurrency
	var wg sync.WaitGroup
	for i := uint64(2); i < toPower; i++ {
		// splits by power of 2 (remnant of old method of checking progress)
		// Can change, but likely won't affect performance
		wg.Add(1)
		start := (uint64(2) << (i - 2)) + 1
		stop := uint64(2) << (i - 1)
		go c.chunkSplitter(start, stop, &wg)
	}
	wg.Wait()
	return true
}

func (c collatzCalculator) chunkSplitter(start uint64, stop uint64, wg *sync.WaitGroup) {
	// break into chunks because the above func does by power of 2
	start = start | 1 // start at odd number
	difference := stop - start
	// min size is 10000 (arbitrary);
	// difference/14 is optimal (for my PC's Ryzen 3900x)
	maxChunkSize := uint64(math.Max(float64(difference/14), float64(10000)))
	cursor := start
	var wg2 sync.WaitGroup
	for cursor < stop { // work through assigned numbers and assign them to threads
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
}

func (c collatzCalculator) calculateChunk(start uint64, stop uint64, wg *sync.WaitGroup) {
	// loop through and actually do the work
	// += 2 because odd numbers only
	for i := start; i < stop; i += 2 {
		c.Test(i)
	}
	wg.Done()
}

func (c collatzCalculator) Test(number uint64) {
	start := number
	// skip all numbers below starting N because run in a loop, this program
	// will test all numbers sequentially, making testing below N redundant
	// and run individually, TestAndPrint will go through all the steps
	for number > start {
		number = c.GetNextCollatzNumber(number)
	}
}

func (c collatzCalculator) TestAndPrint(number uint64) { // print whole path to 1
	// NOT guaranteed to print in order (almost certainly won't)
	// print "N" to file
	c.Printer <- fmt.Sprintf("%d", number)
	for number > 1 {
		// print next number ",N" to file
		number = c.GetNextCollatzNumber(number)
		c.Printer <- fmt.Sprintf(",%d", number)
	}
	// print line break before moving on
	c.Printer <- fmt.Sprintf(",\n")
}

func (c collatzCalculator) GetSteps(number uint64) (steps uint64) {
	// count steps to get to 1
	steps = 0
	for number > 1 {
		steps++
		number = c.GetNextCollatzNumber(number)
	}
	return steps
}

func (c collatzCalculator) GetNextCollatzNumber(number uint64) uint64 {
	// bitshifting for performance (results may vary)
	if number%2 == 0 {
		return number >> 1 // == x2
	}
	// (3n-1)/(2) == (4n-n-1)/(2) == ((4n)-(n-1))/(2) == (2n)-((n-1)/2)
	// These numbers are always odd, so the one's position is always == 1
	// bitshifting drops the one's place, making subtracting 1 unnecessary
	// leaving us with: (2n)-(n/2)
	// Bitshifting left == x2, bitshifting right == /2
	// therefore:
	return (number << 1) - (number >> 1)
}

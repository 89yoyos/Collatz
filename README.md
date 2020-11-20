# Collatz

This program was created to run Collatz Conjecture tests concurrently with some optimisations. It can run silently or print the results. Code is self documenting for the most part. Simple `go get` this repo, import `github.com/89yoyos/Collatz` into your project, and try this:
```
func main() {
	c := Collatz.NewCollatzCalculator()
	c.BenchmarkSequential(24)
}
```
This is then benchmark we used to test its speed. It tests every number from 1 to `2^24` (hence the `24`). It does not print or do anything fancy; you'll need to implement those.

---
This program was written by Alyx Green in November 2020 for Evan Goff's Paper on the Collatz Conjecture.
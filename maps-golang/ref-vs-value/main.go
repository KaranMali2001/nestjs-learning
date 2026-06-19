package main

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

// Job is a "little big" but realistic record — the kind of struct you actually
// pass around a server (modelled on a pipeline job). 5 strings + 5 int64 ≈ 120 B.
//
// This mirrors the boot.dev experiment exactly: one constructor returns the
// struct BY VALUE, the other returns a POINTER. The only difference between the
// two paths is where the struct ends up living (stack vs heap). We measure perf
// INLINE here (runtime.MemStats) instead of using the testing framework, so the
// whole thing runs with a plain `go run ./ref-vs-value`.
type Job struct {
	ID        string
	Pipeline  string
	Name      string
	Command   string
	Status    string
	Attempts  int64
	ExitCode  int64
	CreatedAt int64
	StartedAt int64
	Duration  int64
}

// newJob returns the struct BY VALUE. The string fields are literals (they live
// in the binary's read-only data, NOT the heap), so this constructor does ZERO
// heap allocations — the Job can stay on the caller's stack.
func newJob(i int) Job {
	return Job{
		ID:        "job-7f3a91",
		Pipeline:  "etl-pipeline",
		Name:      "nightly-export",
		Command:   "python export.py --full",
		Status:    "RUNNING",
		Attempts:  int64(i),
		ExitCode:  0,
		CreatedAt: int64(i),
		StartedAt: int64(i + 1),
		Duration:  int64(i + 2),
	}
}

// newJobPtr returns a POINTER to a freshly-built struct. Returning &local forces
// the struct onto the HEAP (escape analysis can't keep it on a stack frame that
// is about to be destroyed) → exactly 1 allocation per call. This is boot.dev's
// newDataPtr with a realistic struct.
func newJobPtr(i int) *Job {
	j := newJob(i)
	return &j
}

// These globals are SINKS. Assigning the result of each call here stops the
// compiler from proving the loop's work is unused and deleting it (the same job
// b.Loop() does automatically in a real benchmark).
var (
	sinkVal Job
	sinkPtr *Job
)

// ---- TEST 2 helpers: a STORED job, updated in place ----
//
// updateValueMap must COPY the whole 120 B struct OUT of the map, modify the
// copy, then COPY it back IN — because map values aren't addressable, so
// `m[id].Attempts++` won't even compile (your day-1 error). Two full copies per
// update.
func updateValueMap(m map[string]Job, id string) {
	j := m[id]   // copy out (120 B)
	j.Attempts++ // modify the copy
	m[id] = j    // copy back in (120 B)
}

// updatePtrMap mutates the heap object in place — no copy, and the natural
// syntax just works because m[id] is a pointer (its target IS addressable).
func updatePtrMap(m map[string]*Job, id string) {
	m[id].Attempts++
}

// measure is a poor-man's benchmark: time a fixed number of iterations and diff
// the allocator counters around the loop. Less precise than testing.B (which
// auto-calibrates the iteration count and de-noises across runs) — but it shows
// what a benchmark does under the hood, and lets us just `go run` the file.
func measure(label string, iters int, f func(i int)) {
	runtime.GC() // clean slate so earlier garbage doesn't skew the counters
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	for i := 0; i < iters; i++ {
		f(i)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&after)
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(iters)
	allocsPerOp := float64(after.Mallocs-before.Mallocs) / float64(iters)
	bytesPerOp := float64(after.TotalAlloc-before.TotalAlloc) / float64(iters)

	fmt.Printf("%-9s %8.2f ns/op   %6.3f allocs/op   %7.1f B/op\n",
		label, nsPerOp, allocsPerOp, bytesPerOp)
}

func main() {
	const iters = 50_000_000

	fmt.Printf("Job struct size: %d B\n", unsafe.Sizeof(Job{}))
	fmt.Printf("iterations:      %d\n\n", iters)

	fmt.Println("TEST 1 — construct & discard (unfair to pointer: pays alloc, no reuse)")
	measure("value", iters, func(i int) { sinkVal = newJob(i) })
	measure("pointer", iters, func(i int) { sinkPtr = newJobPtr(i) })

	// ---- TEST 2 — allocate ONCE, update many times (your map[string]*JobState) ----
	// Setup happens BEFORE measure() opens its counter window, so the one-time
	// allocations don't count against the loop. Both loops should be 0 allocs.
	const id = "job-7f3a91"
	valueMap := map[string]Job{id: newJob(0)}
	ptrMap := map[string]*Job{id: newJobPtr(0)} // 1 alloc, ONCE, here — not per-op

	fmt.Println("\nTEST 2 — allocate once, update N times (fair / real-world)")
	measure("value-map", iters, func(i int) { updateValueMap(valueMap, id) })
	measure("ptr-map", iters, func(i int) { updatePtrMap(ptrMap, id) })

	// Print final state so the updates can't be optimized away.
	fmt.Printf("\nfinal Attempts — value-map: %d, ptr-map: %d\n",
		valueMap[id].Attempts, ptrMap[id].Attempts)
}

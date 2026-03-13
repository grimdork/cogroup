# cogroup

cogroup is a tiny utility that exposes a WaitGroup-like API with a readable
counter. The default exported CoGroup type uses an atomic counter for the
fast path and a sync.WaitGroup for waiting.

Design notes

- The counter is implemented with atomic operations for minimal contention on
  the fast path (Count and Add). Benchmarks show this is measurably cheaper
  than an RWMutex-backed counter in uncontended and common scenarios.
- Values of CoGroup must NOT be copied after first use. The type contains
  sync primitives and copying it can lead to subtle bugs; `go vet` will
  detect such copies when the noCopy sentinel is embedded.

Usage

- Use CoGroup like a sync.WaitGroup: call Add(n), spawn goroutines that call
  Done(), and use Wait() to block until the counter reaches zero. Call
  Count() to read the current counter value.

Good example

This is the recommended pattern: create the CoGroup, Add before spawning
workers, have workers call Done(), and Wait for completion.

```go
var cg cogroup.CoGroup
cg.Add(3)
for i := 0; i < 3; i++ {
    go func() {
        defer cg.Done()
        // work
    }()
}
// optional: read progress
fmt.Println("workers remaining:", cg.Count())
cg.Wait()
```

Bad example — copying (what not to do)

Copying a CoGroup after it has been used can result in two independent values
containing sync primitives; this is a footgun and may lead to incorrect
behaviour or race conditions. `go vet` will warn when it detects copies of
types embedding the noCopy sentinel.

```go
var cg cogroup.CoGroup
cg.Add(1)
// BAD: this copies the internal state
cg2 := cg
go func() { cg.Done() }()
// cg2.Wait() waits on the copy, not the original — bug
cg2.Wait() // may block forever or behave unexpectedly
```

CI

This repository includes a GitHub Actions workflow that runs `go test`, `go
vet` and `staticcheck` on push/pull requests.

Benchmarking

See the benchmarks in waitgroup_test.go. I've added heavier workloads to the
benchmarks (1000 loop iterations per operation) to give more useful numbers on
CI runners and slower machines.

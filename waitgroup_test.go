// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cogroup_test

import (
	"testing"

	"github.com/grimdork/cogroup"
)

func testCoGroup(t *testing.T, co1 *cogroup.CoGroup, co2 *cogroup.CoGroup) {
	n := 16
	co1.Add(n)
	co2.Add(n)
	exited := make(chan bool, n)
	for i := 0; i != n; i++ {
		go func() {
			co1.Done()
			co2.Wait()
			exited <- true
		}()
	}
	co1.Wait()
	for i := 0; i != n; i++ {
		select {
		case <-exited:
			t.Fatal("CoGroup released group too soon")
		default:
		}
		co2.Done()
		t.Logf("Counter 1 at %d", co2.Count())
	}
	for i := 0; i != n; i++ {
		<-exited // Will block if barrier fails to unlock someone.
	}
}

func TestCoGroup(t *testing.T) {
	co1 := &cogroup.CoGroup{}
	co2 := &cogroup.CoGroup{}
	testCoGroup(t, co1, co2)
}

func BenchmarkCoGroupUncontended(b *testing.B) {
	type PaddedCoGroup struct {
		cogroup.CoGroup
		pad [128]uint8
	}
	b.RunParallel(func(pb *testing.PB) {
		var wg PaddedCoGroup
		// reference the padding to avoid staticcheck U1000 "unused field" warning
		_ = wg.pad
		for pb.Next() {
			wg.Add(1)
			wg.Done()
			wg.Wait()
		}
	})
}

func benchmarkCoGroupAddDone(b *testing.B, localWork int) {
	var co cogroup.CoGroup
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			co.Add(1)
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
			co.Done()
		}
		_ = foo
	})
}

func BenchmarkCoGroupAddDone(b *testing.B) {
	benchmarkCoGroupAddDone(b, 0)
}

func BenchmarkCoGroupAddDoneWork(b *testing.B) {
	benchmarkCoGroupAddDone(b, 100)
}

// Heavier work to make benchmarks run longer and give better resolution for
// comparison on slower machines or CI runners.
func BenchmarkCoGroupAddDoneHeavy(b *testing.B) {
	benchmarkCoGroupAddDone(b, 1000)
}

func benchmarkCoGroupWait(b *testing.B, localWork int) {
	var co cogroup.CoGroup
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			co.Wait()
			for i := 0; i < localWork; i++ {
				foo *= 2
				foo /= 2
			}
		}
		_ = foo
	})
}

func BenchmarkCoGroupWait(b *testing.B) {
	benchmarkCoGroupWait(b, 0)
}

func BenchmarkCoGroupWaitWork(b *testing.B) {
	benchmarkCoGroupWait(b, 100)
}

func BenchmarkCoGroupWaitWorkHeavy(b *testing.B) {
	benchmarkCoGroupWait(b, 1000)
}

func BenchmarkWaitGroupActuallyWait(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var co cogroup.CoGroup
			co.Add(1)
			go func() {
				co.Done()
			}()
			co.Wait()
		}
	})
}

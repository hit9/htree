// Copyright 2016 Chao Wang <hit9@icloud.com>.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"github.com/hit9/htree"
	"math/rand"
	"runtime"
)

var (
	size = flag.Int("size", 1000000, "size of the tree to build")
)

func main() {
	flag.Parse()
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	before := stats.Alloc
	t := htree.New()
	for i := 0; i < *size; i++ {
		t.Put(htree.Uint32(rand.Uint32()))
	}
	runtime.GC()
	runtime.ReadMemStats(&stats)
	after := stats.Alloc
	total := float64(after - before)
	fmt.Printf("%5d entry %9.1f B %5.1f B/entry", *size, total, total/float64(*size))
	if t.Len() > 0 { // Make sure sl won't be gc
		fmt.Printf("\n")
	}
}

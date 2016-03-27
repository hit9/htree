// Copyright 2016 Chao Wang <hit9@icloud.com>.

// +build ignore

package main

import (
	"fmt"
	"github.com/hit9/htree"
)

// Item implements htree.Item.
type Item struct {
	key   uint32
	value string
}

// Key returns the item key.
func (item Item) Key() uint32 {
	return item.key
}

func main() {
	t := htree.New()
	// Add an item.
	item := t.Put(Item{123, "data1"})
	// Get an item.
	item = t.Get(Item{key: 123})
	fmt.Println(item)
}

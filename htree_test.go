// Copyright 2016 Chao Wang <hit9@icloud.com>.

package htree

import (
	"math/rand"
	"runtime"
	"testing"
)

// Must asserts the given value is True for testing.
func Must(t *testing.T, v bool) {
	if !v {
		_, fileName, line, _ := runtime.Caller(1)
		t.Errorf("\n unexcepted: %s:%d", fileName, line)
	}
}

func TestPrimesLargerThanUint32(t *testing.T) {
	s := uint64(1)
	for i := 0; i < len(primes); i++ {
		s *= uint64(primes[i])
	}
	Must(t, s > uint64(^uint32(0)))
}

func TestTreeInside(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	n1_0 := tree.root.children[0]
	n1_1 := tree.root.children[1]
	n2_0_0 := n1_0.children[0]
	n2_0_1 := n1_0.children[1]
	n2_0_2 := n1_0.children[2]
	n2_1_0 := n1_1.children[0]
	n2_1_1 := n1_1.children[1]
	n2_1_2 := n1_1.children[2]
	Must(t, n1_0.item == Uint32(0))
	Must(t, n1_1.item == Uint32(1))
	Must(t, n2_0_0.item == Uint32(6))
	Must(t, n2_0_1.item == Uint32(4))
	Must(t, n2_0_2.item == Uint32(2))
	Must(t, n2_1_0.item == Uint32(3))
	Must(t, n2_1_1.item == Uint32(7))
	Must(t, n2_1_2.item == Uint32(5))
	Must(t, len(n2_0_2.children) == 1)
	Must(t, n2_0_2.children[0].item == Uint32(8))
	Must(t, len(n2_1_0.children) == 1)
	Must(t, n2_1_0.children[0].item == Uint32(9))
	Must(t, n1_0.remainder == 0)
	Must(t, n1_1.remainder == 1)
	Must(t, n2_0_0.remainder == 0)
	Must(t, n2_0_1.remainder == 1)
	Must(t, n2_0_2.remainder == 2)
	Must(t, n2_1_0.remainder == 0)
	Must(t, n2_1_1.remainder == 1)
	Must(t, n2_1_2.remainder == 2)
}

func TestPutN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		// Must put
		Must(t, tree.Put(item) != nil)
		// Must get
		Must(t, tree.Get(item) == item)
		// Must len++
		Must(t, tree.Len()+tree.Conflicts() == i+1)
	}
}

func TestPutReuse(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	Must(t, tree.Len() == 10)
	item := Uint32(9)
	Must(t, tree.Put(item) == item)
	Must(t, tree.Conflicts() == 1)
	Must(t, tree.Len() == 10)
}

func TestPutNewNode(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	      |   |
	      8   9       %5
	*/
	tree := New()
	for i := 0; i < 10; i++ {
		tree.Put(Uint32(i))
	}
	Must(t, tree.Len() == 10)
	item := Uint32(10)
	Must(t, tree.Put(item) == item)
	Must(t, tree.Conflicts() == 0)
	Must(t, tree.Len() == 11)
}

func TestGetN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
		// Must get
		Must(t, tree.Get(item) == item)
		// Must cant get
		Must(t, tree.Get(Uint32(n+i)) == nil)
	}
}

func TestDeleteN(t *testing.T) {
	tree := New()
	n := 1024
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
		// Must delete
		Must(t, tree.Delete(item) == item)
		// Must cant delete
		Must(t, tree.Delete(Uint32(n+i)) == nil)
		// Must len--
		Must(t, tree.Len() == 0)
	}
}

func TestDeleteReplace(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	  |    |
	  42   8          %5
	*/
	tree := New()
	for i := 0; i < 9; i++ {
		tree.Put(Uint32(i))
	}
	tree.Put(Uint32(42))
	Must(t, tree.Len() == 10)
	item := Uint32(0)
	// Must delete
	Must(t, tree.Delete(item) == item)
	// Original child must be replaced by new node:42
	Must(t, tree.root.children[0].item == Uint32(42))
	Must(t, tree.root.children[0].remainder == 0)
	Must(t, tree.root.children[0].depth == 1)
	// The children shouldnt be changed
	Must(t, len(tree.root.children[0].children) == 3)
	// Node must be a leaf now.
	leaf := tree.root.children[0].children[0]
	Must(t, len(leaf.children) == 0)
	// Must length--
	Must(t, tree.Len() == 9)
}

func TestDeleteLeaf(t *testing.T) {
	/*
	       root
	     /     \
	    0       1     %2
	   /|\     /|\
	  6 4 2   3 7 5   %3
	*/
	tree := New()
	for i := 0; i < 8; i++ {
		tree.Put(Uint32(i))
	}
	Must(t, tree.Len() == 8)
	item := Uint32(7)
	// Must delete
	Must(t, tree.Delete(item) == item)
	// Must node(1) has 2 nodes now
	Must(t, len(tree.root.children[1].children) == 2)
	// Must length--
	Must(t, tree.Len() == 7)
}

func TestIteratorEmpty(t *testing.T) {
	tree := New()
	i := 0
	iter := tree.NewIterator()
	for iter.Next() {
		i++
	}
	// Must iterates 0 times
	Must(t, i == 0)
}

func TestIteratorOrder(t *testing.T) {
	/*
	      root
	     /    \
	    0      1     %2
	   / \    / \
	  4   2  3   5   %3
	*/
	tree := New()
	tree.Put(Uint32(0))
	tree.Put(Uint32(1))
	tree.Put(Uint32(2))
	tree.Put(Uint32(3))
	tree.Put(Uint32(4))
	tree.Put(Uint32(5))
	iter := tree.NewIterator()
	Must(t, iter.Next() && iter.Item() == Uint32(0))
	Must(t, iter.Next() && iter.Item() == Uint32(4))
	Must(t, iter.Next() && iter.Item() == Uint32(2))
	Must(t, iter.Next() && iter.Item() == Uint32(1))
	Must(t, iter.Next() && iter.Item() == Uint32(3))
	Must(t, iter.Next() && iter.Item() == Uint32(5))
}

func TestIteratorLarge(t *testing.T) {
	tree := New()
	n := 1024 * 10
	for i := 0; i < n; i++ {
		item := Uint32(rand.Uint32())
		tree.Put(item)
	}
	j := 0
	iter := tree.NewIterator()
	for iter.Next() {
		j++
	}
	Must(t, j == tree.Len())
}

func BenchmarkPut(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
}

func BenchmarkGet(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(Uint32(i))
	}
}

func BenchmarkIteratorNext(b *testing.B) {
	t := New()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
	iter := t.NewIterator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Next()
	}
}

func BenchmarkGetLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(Uint32(i))
	}
}

func BenchmarkPutLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Put(Uint32(i))
	}
}

func BenchmarkDeleteLargeTree(b *testing.B) {
	t := New()
	n := 1000 * 1000 // Million
	for i := 0; i < n; i++ {
		t.Put(Uint32(rand.Uint32()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Delete(Uint32(i))
	}
}

// Copyright 2016 Chao Wang <hit9@icloud.com>.

/*

Package htree implements the in-memory hash tree.

Abstract

Hash-Tree is a key-value multi-tree with fast indexing performance and
high space utilization.

Take 10 consecutive prime numbers:

	2, 3, 5, 7, 11, 13, 17, 19, 23, 29

And they can distinguish all uint32 numbers:

	2*3*5*7*11*13*17*19*23*29 > ^uint32(0)

Key insertion steps:

	1. Suppose depth is d, its prime p = Primes[d]
	2. Get the remainder of key divided by p, r = key % p
	3. If r is already taken by another node, go to the next depth.
	4. Otherwise create a new child node and bind r to it.

Example htree:

	ROOT
	  | %2
	  +-- 0: A(12)         12%2=0: A
	  |       | %3
	  |       +-- 0: C(6)  6%2=0 && 6%3=0: C
	  |       |-- 1: D(4)  4%2=0 && 4%3=1: D
	  |       +-- 2: E(8)  8%2=0 && 8%3=2: E
	  +-- 1: B(3)          3%2=1: B
	          | %3
	          +-- 0: F(9)  9%2=1 && 9%3=0: F
	          |-- 1: G(7)  7%2=1 && 7%3=1: G
	          +-- 2: H(11) 11%2=1 && 11%3=2: H

Complexity

Child nodes are stored in an array orderly, and checked by binary-search but
not indexed by remainders, array indexing will result redundancy entries, this
is for less memory usage, with a bit performance loss. A node is created only
when a new key is inserted. So the worst time complexity is O(Sum(lg2~lg29)),
is constant level, and the entries utilization is 100%.

Compare To Map

Although hashtable is very fast with O(1) time complexity, but there is always
about ~25% table entries are unused, because the hash-table load factor is .75.
And this htree is suitable for memory-bounded cases.

HTree is better for local locks if you want a safe container.

Map may need to rehash and resize on insertions.

Goroutine Safety.

No. Lock granularity depends on the use case.

*/
package htree // import "github.com/hit9/htree"

// Item is a single object in the tree.
type Item interface {
	// Key returns an uint32 number to distinguish node with another.
	Key() uint32
}

// Uint32 implements the Item interface.
type Uint32 uint32

// Key returns the htree node key.
func (i Uint32) Key() uint32 {
	return uint32(i)
}

type children []*node

// node is an internel node in the htree.
type node struct {
	item      Item
	depth     int8     // int8 number on [0,10]
	remainder int8     // item.Key()%primes[father.depth]
	children  children // ordered by remainder
}

// HTree is the hash-tree.
type HTree struct {
	root      *node // empty root node
	length    int   // number of nodes
	conflicts int   // number of conflicts
}

// Iterator is an iterator on the htree.
type Iterator struct {
	t       *HTree
	fathers []*node // stack of father node
	indexes []int   // stack of father's index in the brothers
	n       *node   // current node
	i       int     // current index in n's brothers
}

// Prime numbers to build the tree.
var primes = [10]int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}

// modulo returns the remainder after division of key by the prime.
func modulo(key uint32, depth int8) int8 {
	return int8(key % uint32(primes[depth]))
}

// newNode creates a new node.
func newNode(item Item, depth int8, remainder int8) *node {
	// item,depth,remainder won't be rewritten once init.
	return &node{
		item:      item,
		depth:     depth,
		remainder: remainder,
	}
}

// insert a node into the children slice at index i.
func (s *children) insert(i int, n *node) {
	*s = append(*s, nil)
	if i < len(*s) {
		copy((*s)[i+1:], (*s)[i:])
	}
	(*s)[i] = n
}

// delete a node from the children slice at index i.
func (s *children) delete(i int) {
	(*s) = append((*s)[:i], (*s)[i+1:]...)
}

// search child by remainder via binary-search, returns the result
// and left/right positions.
func (s *children) search(r int8) (ok bool, left, right int) {
	right = len(*s) - 1
	for left < right {
		mid := (left + right) >> 1
		child := (*s)[mid]
		if r > child.remainder {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if left == right {
		child := (*s)[left]
		if r == child.remainder {
			ok = true
			return
		}
	}
	return
}

// New creates a new htree.
func New() *HTree {
	return &HTree{root: &node{}}
}

// Len returns the number of nodes in the tree.
func (t *HTree) Len() int { return t.length }

// Conflicts returns the number of conflicts in the tree.
func (t *HTree) Conflicts() int { return t.conflicts }

// get item recursively, nil on not found.
func (t *HTree) get(n *node, item Item) Item {
	r := modulo(item.Key(), n.depth)
	ok, left, _ := n.children.search(r)
	if ok {
		// Get the child with the same remainder.
		child := n.children[left]
		if child.item.Key() == item.Key() {
			// Found.
			return child.item
		}
		// Next depth.
		return t.get(child, item)
	}
	// Not found.
	return nil
}

// put finds item recursively, if the node with given item is
// found, returns it. Otherwise new a node with the item.If the
// depth overflows, nil is returned.
func (t *HTree) put(n *node, item Item) Item {
	r := modulo(item.Key(), n.depth)
	ok, left, right := n.children.search(r)
	if ok {
		// Get the child with the same remainder.
		child := n.children[left]
		if child.item.Key() == item.Key() {
			t.conflicts++
			return child.item // reuse
		}
		// Next depth.
		return t.put(child, item)
	}
	if n.depth >= int8(len(primes)-1) {
		return nil // depth overflows
	}
	// Create a new node.
	child := newNode(item, n.depth+1, r)
	if len(n.children) == 0 || (right == len(n.children)-1 &&
		r >= n.children[right].remainder) {
		n.children = append(n.children, child)
	} else {
		n.children.insert(right, child)
	}
	t.length++
	return child.item
}

// delete finds node by item recursively, if found, deletes it and
// returns the item, else nil.
func (t *HTree) delete(n *node, item Item) Item {
	r := modulo(item.Key(), n.depth)
	ok, left, _ := n.children.search(r)
	if ok {
		// Get the child with the same remaider.
		child := n.children[left]
		if child.item.Key() == item.Key() {
			if len(child.children) == 0 {
				// Delete child directly.
				n.children.delete(left)
			} else {
				// Find the leaf on this branch.
				father := child
				leaf := father.children[0]
				for {
					if len(leaf.children) == 0 {
						break
					}
					father = leaf
					leaf = father.children[0]
				}
				// Replace child with new node.
				father.children.delete(0)
				n.children[left] = newNode(leaf.item, child.depth, child.remainder)
				n.children[left].children = child.children
			}
			t.length--
			return child.item
		}
		return t.delete(child, item)
	}
	return nil
}

// Get item from htree, nil if not found.
func (t *HTree) Get(item Item) Item {
	return t.get(t.root, item)
}

// Put item into htree and returns the item. If the item already in the
/// tree, return it, else new a node with the given item and return this
// item. If the depth overflows, nil is returned.
func (t *HTree) Put(item Item) Item {
	return t.put(t.root, item)
}

// Delete item from htree and returns the item, nil on not found.
func (t *HTree) Delete(item Item) Item {
	return t.delete(t.root, item)
}

// NewIterator returns a new iterator on this htree.
func (t *HTree) NewIterator() *Iterator {
	return &Iterator{n: t.root, i: 0, t: t}
}

// Next seeks the iterator to next.
// Iteration order sample:
//
// 	      root
// 	     /    \
// 	    0      1     %2
// 	   / \    / \
// 	  4   2  3   5   %3
//
// Order: 0 -> 4 -> 2 -> 1 -> 3 -> 5
func (iter *Iterator) Next() bool {
	if len(iter.n.children) > 0 {
		// Push stack
		iter.fathers = append(iter.fathers, iter.n)
		iter.indexes = append(iter.indexes, iter.i)
		iter.n = iter.n.children[0]
		iter.i = 0
		return true
	}
	for len(iter.fathers) > 0 {
		l := len(iter.fathers)
		father := iter.fathers[l-1]
		if iter.i < len(father.children)-1 {
			iter.i++
			iter.n = father.children[iter.i]
			return true
		}
		// Pop stack
		iter.fathers, father = iter.fathers[:l-1], iter.fathers[l-1]
		iter.indexes, iter.i = iter.indexes[:l-1], iter.indexes[l-1]
	}
	return false
}

// Item returns the current item.
func (iter *Iterator) Item() Item {
	return iter.n.item
}

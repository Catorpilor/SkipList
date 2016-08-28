package skiplist

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	maxLevel    = 32
	probability = 2 // 1/2

	// coins
	COINHEADS = 1
	COINTAILS = 0

	//compares
	COMPARESAME  = 0
	COMPAREGREAT = 1
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Node struct {
	item interface{}
	next []*Node //multiple level
	prev *Node   //just for level 0
}

// Compare returns
// 1 - if value "a" is greater than value "b"
// 0 - then are the same
// -1 - if value "a" is less than value "b"
type Compare func(a, b interface{}) int

type SkipList struct {
	header  *Node
	len     int
	level   int // current level
	compare Compare
	mu      sync.RWMutex
}

// NewList takes a compare func to make the interface comparable
func NewList(compare Compare) *SkipList {
	head := &Node{
		next: make([]*Node, maxLevel),
	}
	return &SkipList{
		len:     0,
		level:   0,
		compare: compare,
		header:  head,
	}
}

// flipCoin returns 1 if it is heads, 0 otherwise
func flipCoin() int {
	return r.Intn(probability)
}

// "Fix the dice" strategy
// Generate a level that is at most 1 higher than current max
func (l *SkipList) randLevel() int {
	var level int
	for level = 0; r.Int31n(probability) == 1 && level < maxLevel; level++ {
		if level > l.level {
			break
		}
	}

	return level
}

// InsertV2 is the topdown approach
// No need to store the going down node for future link update
func (l *SkipList) InsertV2(item interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	// flip coins first to find out which level the Node{item} will be
	level := l.randLevel()

	// if higher update the SkipList's level
	if level > l.level {
		l.level = level
	}

	inserted := false
	node := &Node{item: item, next: make([]*Node, level+1)}
	x := l.header
	for i := l.level; i >= 0; i-- {
		for x.next[i] != nil {
			// Found larger next item, slot found.
			compare := l.compare(x.next[i].item, item)

			// Fail on duplicate.
			if compare == COMPARESAME {
				return false
			}

			// If next item is greater than insert subject.
			// Found correct slot.
			if compare == 1 {
				break
			}

			// No match found, move forward.
			x = x.next[i]
		}

		// Keep moving down
		if i > level {
			continue
		}

		node.next[i] = x.next[i]
		// On level 0, link up previous nodes.
		if i == 0 {
			node.prev = x
			if x.next[i] != nil {
				x.next[i].prev = node
			}
		}
		x.next[i] = node

		if !inserted {
			inserted = true
		}

		// No match found on this level, move down on next iteration.
	}

	if inserted {
		l.len++
	}
	return inserted
}

// Insert x insert into sk returns true if success,otherwise returns false
// Insert is a bottom up approach, which described by the algorithm.
func (l *SkipList) Insert(item interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Step 1 search from the top lelf
	knots := make(map[int]*Node, l.level+1) //knots stores going down node on each level
	x := l.header
	node := &Node{
		next: make([]*Node, l.level+2), //allocate current level + 2 cause l.level start at 0 makes it only 1 level higher than the current level
		item: item,
	}
	for i := l.level; i >= 0; i-- {
		for x.next[i] != nil {
			// compare
			compare := l.compare(x.next[i].item, item)

			// duplicate item
			if compare == COMPARESAME {
				return false
			}

			if compare == COMPAREGREAT {
				// insert into current slot
				break
			}

			// less than item move forward
			x = x.next[i]
		}

		if i != 0 {
			knots[i] = x //going down store node in knots
			continue
		}

		// bottom linked list
		node.next[i] = x.next[i]
		node.prev = x
		if x.next[i] != nil {
			x.next[i].prev = node
		}
		x.next[i] = node

		// increase count
		l.len++
	}

	// flip coins
	var level int
	for level = 1; flipCoin() == COINHEADS && level < maxLevel; level++ {
		if level > l.level {
			l.header.next[level] = node // the first node on the new level,just link it with header
			l.level = level             // update list's level
			break
		}

		// promote to next level, link with knot[level]
		node.next[level] = knots[level].next[level]
		knots[level].next[level] = node
	}

	return true
}

// Exists check if item in list returns true otherwise returns false
func (l *SkipList) Exists(item interface{}) bool {
	// search from top left
	l.mu.Lock()
	defer l.mu.Unlock()
	v := l.header

	for i := l.level; i >= 0; i-- {
		for v.next[i] != nil {

			// compare item
			compare := l.compare(v.next[i].item, item)

			if compare == COMPARESAME {
				return true
			}

			if compare == COMPAREGREAT {
				break
			}

			// less, move forward
			v = v.next[i]
		}

		// no match on this level, going down
	}

	return false
}

// Delete delete item from skip lists
func (l *SkipList) Delete(item interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	found := false
	var v, prev *Node
	v = l.header
	var i int
	for i = l.level; i >= 0; i-- {
		for v.next[i] != nil {
			// compare
			compare := l.compare(v.next[i].item, item)
			if compare == COMPARESAME {
				found = true
				break
			}
			if compare == COMPAREGREAT {
				break
			}
			v = v.next[i]
		}

		if !found {
			continue
		}
		prev = v
		next := v.next[i].next[i]
		prev.next[i] = next

		if i == 0 {
			if next != nil {
				next.prev = prev
			}
		}

		// Determine if we have deleted the max item in current level.
		// If so, reduce max level of list.
		if next == nil && prev == l.header && i > 0 {
			l.level = i - 1
		}
	}

	if found {
		l.len--
	}

	return found
}

// Len returns the length of skip lists
func (l *SkipList) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.len
}

// Print just for debug
func (l *SkipList) Print() {
	x := l.header
	fmt.Printf("header: %#v\n", x)
	for x.next[0] != nil {
		fmt.Printf("element: %#v, address: %p\n", x.next[0], x.next[0])
		x = x.next[0]
	}
}

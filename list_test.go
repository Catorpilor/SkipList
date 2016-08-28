package skiplist

import (
	"math/rand"
	"sort"
	"testing"
)

func compareIntAsc(a interface{}, b interface{}) int {
	aa := a.(int)
	bb := b.(int)

	if aa < bb {
		return -1
	}

	if aa > bb {
		return 1
	}

	return 0
}

func compareIntDesc(a interface{}, b interface{}) int {
	r := compareIntAsc(a, b)
	if r == 1 {
		return -1
	}

	if r == -1 {
		return 1
	}

	return r
}

func TestInsertDuplicate(t *testing.T) {
	l := NewList(compareIntAsc)
	a := 1

	if ok := l.Insert(a); !ok {
		t.Fatalf("Insert failed")
	}

	if l.Len() != 1 {
		t.Fatalf("Length mismatch, exp=%d, act=%d", 1, l.Len())
	}

	if ok := l.Insert(a); ok {
		t.Fatalf("Duplicate insert should fail")
	}

	if l.Len() != 1 {
		t.Fatalf("Length mismatch, exp=%d, act=%d", 1, l.Len())
	}
}

func TestExists(t *testing.T) {
	l := NewList(compareIntAsc)
	a := 1

	if l.Exists(a) {
		t.Fatalf("Unexpected item found")
	}

	if !l.Insert(a) {
		t.Fatalf("Insert failed")
	}

	if !l.Exists(a) {
		t.Fatalf("Expected item to be found")
	}
}

func TestInsertAndDelete(t *testing.T) {
	type testInput struct {
		items []int
		asc   bool
	}
	tests := []testInput{
		{[]int{1}, true},
		{[]int{1}, false},

		{[]int{1, 2}, true},
		{[]int{1, 2}, false},

		{[]int{2, 1}, true},
		{[]int{2, 1}, false},
		{[]int{2, 1, 3, 5, 4, 7, 6}, true},
		{[]int{2, 1, 3, 5, 4, 7, 6}, false},
	}
	items := make([]int, 256)
	for i := 0; i < len(items); i++ {
		items[i] = i
	}
	tests = append(tests, testInput{items, true})
	tests = append(tests, testInput{items, false})

	items = rand.Perm(4096)
	tests = append(tests, testInput{items, true})
	tests = append(tests, testInput{items, false})
	for _, tt := range tests {
		leveledUp := false
		for {
			l := NewList(compareIntAsc)
			if !tt.asc {
				l = NewList(compareIntDesc)
			}

			x := l.header
			for k, v := range tt.items {
				if !l.Insert(v) {
					t.Fatalf("Insert failed")
				}

				if x.prev != nil || x.item != nil {
					t.Fatalf("Invalid list head")
				}

				if l.Len() != k+1 {
					t.Fatalf("List length mismatch, exp=%d, act=%d", k+1, l.Len())
				}
			}

			if l.level > 0 {
				leveledUp = true
			}

			//l.Print()

			if l.Len() != len(tt.items) {
				t.Fatalf("Length mismatch, exp=%d, act=%d", len(tt.items), l.Len())
			}

			sorted := make(sort.IntSlice, len(tt.items))
			copy(sorted, tt.items)
			if tt.asc {
				sorted.Sort()
			} else {
				sort.Sort(sort.Reverse(sorted))
			}

			// Verify the items are in order on level 0 in BOTH directions.
			for i := 0; i < len(tt.items); i++ {
				if x.next[0].item != sorted[i] {
					t.Fatalf("Next item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].item, sorted[i])
				}

				if i > 0 && x.next[0].prev.item != sorted[i-1] {
					t.Fatalf("Previous item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].prev.item, sorted[i-1])
				}

				x = x.next[0]
			}

			if x.next[0] != nil {
				t.Fatalf("End of list mismatch, next=%+v", x.next[0])
			}

			for k, v := range tt.items {
				if !l.Delete(v) {
					t.Fatalf("Delete failed")
				}

				if l.Len() != (len(tt.items) - k - 1) {
					t.Fatalf("List len mismatch, exp=%d, act=%d", (len(tt.items) - k - 1), l.Len())
				}

				idx := -1
				for kk, vv := range sorted {
					if vv == v {
						idx = kk
						break
					}
				}

				if idx == -1 {
					t.Fatalf("Unable to find value in sorted items")
				}

				// Delete from sortered list.
				sorted = append(sorted[:idx], sorted[idx+1:]...)

				x := l.header
				if len(sorted) == 0 {
					if x.next[0] != nil {
						t.Fatalf("Expected end of list.next to be nil, next=%+v", x.next[0])
					}
				}

				for i := 0; i < len(sorted); i++ {
					if x.next[0].item != sorted[i] {
						t.Fatalf("Next item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].item, sorted[i])
					}

					if i > 0 && x.next[0].prev.item != sorted[i-1] {
						t.Fatalf("Previous item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].prev.item, sorted[i-1])
					}

					x = x.next[0]
				}
			}

			if leveledUp {
				break
			}
		} // End for {}
	} // End range tests
}
func TestInsertV2AndDelete(t *testing.T) {
	type testInput struct {
		items []int
		asc   bool
	}
	tests := []testInput{
		{[]int{1}, true},
		{[]int{1}, false},

		{[]int{1, 2}, true},
		{[]int{1, 2}, false},

		{[]int{2, 1}, true},
		{[]int{2, 1}, false},

		{[]int{2, 1, 3}, true},
		{[]int{2, 1, 3}, false},
	}
	items := make([]int, 256)
	for i := 0; i < len(items); i++ {
		items[i] = i
	}
	tests = append(tests, testInput{items, true})
	tests = append(tests, testInput{items, false})

	items = rand.Perm(4096)
	tests = append(tests, testInput{items, true})
	tests = append(tests, testInput{items, false})
	for _, tt := range tests {
		leveledUp := false
		for {
			l := NewList(compareIntAsc)
			if !tt.asc {
				l = NewList(compareIntDesc)
			}

			x := l.header
			for k, v := range tt.items {
				if !l.InsertV2(v) {
					t.Fatalf("Insert failed")
				}

				if x.prev != nil || x.item != nil {
					t.Fatalf("Invalid list head")
				}

				if l.Len() != k+1 {
					t.Fatalf("List length mismatch, exp=%d, act=%d", k+1, l.Len())
				}
			}

			if l.level > 0 {
				leveledUp = true
			}

			if l.Len() != len(tt.items) {
				t.Fatalf("Length mismatch, exp=%d, act=%d", len(tt.items), l.Len())
			}

			sorted := make(sort.IntSlice, len(tt.items))
			copy(sorted, tt.items)
			if tt.asc {
				sorted.Sort()
			} else {
				sort.Sort(sort.Reverse(sorted))
			}

			// Verify the items are in order on level 0 in BOTH directions.
			for i := 0; i < len(tt.items); i++ {
				if x.next[0].item != sorted[i] {
					t.Fatalf("Next item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].item, sorted[i])
				}

				if i > 0 && x.next[0].prev.item != sorted[i-1] {
					t.Fatalf("Previous item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].prev.item, sorted[i-1])
				}

				x = x.next[0]
			}

			if x.next[0] != nil {
				t.Fatalf("End of list mismatch, next=%+v", x.next[0])
			}

			for k, v := range tt.items {
				if !l.Delete(v) {
					t.Fatalf("Delete failed")
				}

				if l.Len() != (len(tt.items) - k - 1) {
					t.Fatalf("List len mismatch, exp=%d, act=%d", (len(tt.items) - k - 1), l.Len())
				}

				idx := -1
				for kk, vv := range sorted {
					if vv == v {
						idx = kk
						break
					}
				}

				if idx == -1 {
					t.Fatalf("Unable to find value in sorted items")
				}

				// Delete from sortered list.
				sorted = append(sorted[:idx], sorted[idx+1:]...)

				x := l.header
				if len(sorted) == 0 {
					if x.next[0] != nil {
						t.Fatalf("Expected end of list.next to be nil, next=%+v", x.next[0])
					}
				}

				for i := 0; i < len(sorted); i++ {
					if x.next[0].item != sorted[i] {
						t.Fatalf("Next item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].item, sorted[i])
					}

					if i > 0 && x.next[0].prev.item != sorted[i-1] {
						t.Fatalf("Previous item mismatch, i=%d, a=%+v, b=%+v", i, x.next[0].prev.item, sorted[i-1])
					}

					x = x.next[0]
				}
			}

			if leveledUp {
				break
			}
		} // End for {}
	} // End range tests
}

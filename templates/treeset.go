// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package treeset implements a tree backed by a red-black tree.
//
// Structure is not thread safe.
//
// Reference: http://en.wikipedia.org/wiki/Set_%28abstract_data_type%29
package treeset

import (
	"encoding/json"
	"fmt"
	rbt "github.com/eosspark/container/trees/redblacktree"
	"github.com/eosspark/container/utils"
	"strings"
)

//func assertSetImplementation() {
//	var _ sets.Set = (*Set)(nil)
//}
type Comparable interface {
	Compare(c Comparable) int
}

// template type Set(V)
type V = Comparable

// Set holds elements in a red-black tree
type Set struct {
	tree      *rbt.Tree
}

var itemExists = struct{}{}

// NewWith instantiates a new empty set with the custom comparator.

func New(Value ...V) *Set {
	return &Set{tree: rbt.NewWith(func(a, b interface{}) int {
		return a.(V).Compare(b.(V))
	})}
}
func NewWith(comparator utils.Comparator, values ...V) *Set {
	set := &Set{tree: rbt.NewWith(comparator)}
	if len(values) > 0 {
		set.Add(values...)
	}
	return set
}


func CopyFrom(ts *Set) *Set {
	return &Set{tree: rbt.CopyFrom(ts.tree)}
}

func SetIntersection(a *Set, b *Set, callback func(elem V)) {
	aIterator := a.Iterator()
	bIterator := b.Iterator()

	if !aIterator.First() || !bIterator.First() {
		return
	}

	comparator := a.GetComparator()

	for aHasNext, bHasNext := true, true; aHasNext && bHasNext; {
		comp := comparator(aIterator.Value(), bIterator.Value())
		switch {
		case comp > 0:
			bHasNext = bIterator.Next()
		case comp < 0:
			aHasNext = aIterator.Next()
		default:
			callback(aIterator.Value())
			aHasNext = aIterator.Next()
			bHasNext = bIterator.Next()
		}
	}
}

func (set *Set) GetComparator() utils.Comparator {
	return set.tree.Comparator
}

// Add adds the item one to the set.Returns false and the interface if it already exists
func (set *Set) AddItem(item V) (bool, V) {
	opt, k, _ := set.tree.PutItem(item, itemExists)
	return opt, k.(V)
}

// Add adds the items (one or more) to the set.
func (set *Set) Add(items ...V) {
	for _, item := range items {
		set.tree.Put(item, itemExists)
	}
}

// Remove removes the items (one or more) from the set.
func (set *Set) Remove(items ...V) {
	for _, item := range items {
		set.tree.Remove(item)
	}
}

// Contains checks weather items (one or more) are present in the set.
// All items have to be present in the set for the method to return true.
// Returns true if no arguments are passed at all, i.e. set is always superset of empty set.
func (set *Set) Contains(items ...V) bool {
	for _, item := range items {
		if _, contains := set.tree.Get(item); !contains {
			return false
		}
	}
	return true
}

// Empty returns true if set does not contain any elements.
func (set *Set) Empty() bool {
	return set.tree.Size() == 0
}

// Size returns number of elements within the set.
func (set *Set) Size() int {
	return set.tree.Size()
}

// Clear clears all values in the set.
func (set *Set) Clear() {
	set.tree.Clear()
}

// Values returns all items in the set.
func (set *Set) Values() []V {
	keys := make([]V, set.Size())
	it := set.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Value()
	}
	return keys
}

// String returns a string representation of container
func (set *Set) String() string {
	str := "TreeSet\n"
	items := []string{}
	for _, v := range set.tree.Keys() {
		items = append(items, fmt.Sprintf("%v", v))
	}
	str += strings.Join(items, ", ")
	return str
}

// Iterator returns a stateful iterator whose values can be fetched by an index.
type Iterator struct {
	index    int
	iterator rbt.Iterator
	tree     *rbt.Tree
}

// Iterator holding the iterator's state
func (set *Set) Iterator() Iterator {
	return Iterator{index: -1, iterator: set.tree.Iterator(), tree: set.tree}
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's index and value can be retrieved by Index() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (iterator *Iterator) Next() bool {
	if iterator.index < iterator.tree.Size() {
		iterator.index++
	}
	return iterator.iterator.Next()
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (iterator *Iterator) Prev() bool {
	if iterator.index >= 0 {
		iterator.index--
	}
	return iterator.iterator.Prev()
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (iterator *Iterator) Value() V {
	return iterator.iterator.Key().(V)
}

// Index returns the current element's index.
// Does not modify the state of the iterator.
func (iterator *Iterator) Index() int {
	return iterator.index
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (iterator *Iterator) Begin() {
	iterator.index = -1
	iterator.iterator.Begin()
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (iterator *Iterator) End() {
	iterator.index = iterator.tree.Size()
	iterator.iterator.End()
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (iterator *Iterator) First() bool {
	iterator.Begin()
	return iterator.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's index and value can be retrieved by Index() and Value().
// Modifies the state of the iterator.
func (iterator *Iterator) Last() bool {
	iterator.End()
	return iterator.Prev()
}


// Each calls the given function once for each element, passing that element's index and value.
func (set *Set) Each(f func(index int, value V)) {
	iterator := set.Iterator()
	for iterator.Next() {
		f(iterator.Index(), iterator.Value())
	}
}

// Map invokes the given function once for each element and returns a
// container containing the values returned by the given function.
func (set *Set) Map(f func(index int, value V) V) *Set {
	newSet := &Set{tree: rbt.NewWith(set.tree.Comparator)}
	iterator := set.Iterator()
	for iterator.Next() {
		newSet.Add(f(iterator.Index(), iterator.Value()))
	}
	return newSet
}

// Select returns a new container containing all elements for which the given function returns a true value.
func (set *Set) Select(f func(index int, value V) bool) *Set {
	newSet := &Set{tree: rbt.NewWith(set.tree.Comparator)}
	iterator := set.Iterator()
	for iterator.Next() {
		if f(iterator.Index(), iterator.Value()) {
			newSet.Add(iterator.Value())
		}
	}
	return newSet
}

// Any passes each element of the container to the given function and
// returns true if the function ever returns true for any element.
func (set *Set) Any(f func(index int, value V) bool) bool {
	iterator := set.Iterator()
	for iterator.Next() {
		if f(iterator.Index(), iterator.Value()) {
			return true
		}
	}
	return false
}

// All passes each element of the container to the given function and
// returns true if the function returns true for all elements.
func (set *Set) All(f func(index int, value V) bool) bool {
	iterator := set.Iterator()
	for iterator.Next() {
		if !f(iterator.Index(), iterator.Value()) {
			return false
		}
	}
	return true
}

// Find passes each element of the container to the given function and returns
// the first (index,value) for which the function is true or -1,nil otherwise
// if no element matches the criteria.
func (set *Set) Find(f func(index int, value V) bool) (idx int, v V) {
	iterator := set.Iterator()
	for iterator.Next() {
		if f(iterator.Index(), iterator.Value()) {
			return iterator.Index(), iterator.Value()
		}
	}
	return -1, v
}

// ToJSON outputs the JSON representation of the set.
func (set *Set) ToJSON() ([]byte, error) {
	return json.Marshal(set.Values())
}

// FromJSON populates the set from the input JSON representation.
func (set *Set) FromJSON(data []byte) error {
	elements := []V{}
	err := json.Unmarshal(data, &elements)
	if err == nil {
		set.Clear()
		set.Add(elements...)
	}
	return err
}
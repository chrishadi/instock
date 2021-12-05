package toplist

import (
	"testing"
)

const size = 5

var more = func(a, b interface{}) bool { return a.(int) > b.(int) }
var less = func(a, b interface{}) bool { return a.(int) < b.(int) }

func TestEmpty(t *testing.T) {
	t.Skip()
}

func TestAddGivenNewValueWhenListIsNotFullShouldInsertTheValue(t *testing.T) {
	tl := New(size, more)

	if !tl.Add(1) {
		t.Fatal("Expect first add to succeed, but fail")
	}
	if !tl.Add(2) {
		t.Fatal("Expect second add to succeed, but fail")
	}
	list := tl.Elements()
	if list.Len() != 2 {
		t.Fatal("Expect list length to be 1, got", list.Len())
	}
	first := list.Front().Value.(int)
	if first != 2 {
		t.Fatal("Expect 2 as the first element, got", first)
	}
}

func TestAddGivenNewValueIsLessThanOrEqualBottomWhenListIsFullShouldNotInsert(t *testing.T) {
	tl := New(size, more)
	tl.Add(2)
	tl.Add(3)
	tl.Add(4)
	tl.Add(5)
	tl.Add(6)

	if tl.Add(1) {
		t.Error("Expect add value less than bottom to fail, but succeed")
	}
	if tl.Add(2) {
		t.Error("Expect add value equal to bottom to fail, but succeed")
	}
}

func TestAddGivenNewValueIsGreaterThanTopShouldInsertAtTheTop(t *testing.T) {
	tl := New(size, more)
	tl.Add(2)
	tl.Add(1)
	tl.Add(3)

	top := tl.Elements().Front().Value.(int)
	if top != 3 {
		t.Error("Expect 3 at the top, got", top)
	}
}

func TestAddGivenNewValueIsLessThanBottomShouldInsertAtTheBottom(t *testing.T) {
	tl := New(size, more)
	tl.Add(2)
	tl.Add(3)
	tl.Add(1)

	bottom := tl.Elements().Back().Value.(int)
	if bottom != 1 {
		t.Error("Expect 1 at the bottom, got", bottom)
	}
}

func TestAddGivenNewValueIsSomewhereBetweenShouldInsertSomewhereBetween(t *testing.T) {
	tl := New(size, less)
	tl.Add(-4)
	tl.Add(-2)
	tl.Add(-1)
	tl.Add(-3)

	second := tl.Elements().Front().Next().Value.(int)
	if second != -3 {
		t.Error("Expect -3 at second position, got", second)
	}
}

func TestAddGivenNewStockGainIsGreaterThanSomeWhenListIsFullShouldRemoveLast(t *testing.T) {
	tl := New(size, more)
	tl.Add(110)
	tl.Add(109)
	tl.Add(108)
	tl.Add(106)
	tl.Add(105)
	tl.Add(107)

	list := tl.Elements()
	if list.Len() != size {
		t.Errorf("Expect list len to be %d, got %d", size, list.Len())
	}
	bottom := list.Back().Value.(int)
	if bottom != 106 {
		t.Error("Expect 106 at the bottom, got", bottom)
	}
}

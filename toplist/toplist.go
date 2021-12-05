package toplist

import "container/list"

type TopList struct {
	list *list.List
	size int
	more func(a, b interface{}) bool
}

func New(size int, moreFunc func(a, b interface{}) bool) *TopList {
	return &TopList{
		list: list.New(),
		size: size,
		more: moreFunc,
	}
}

func (tl *TopList) Add(v interface{}) bool {
	list := tl.Elements()
	if list.Len() == tl.size {
		back := list.Back()
		if !tl.more(v, back.Value) {
			return false
		} else {
			list.Remove(back)
		}
	}

	front := list.Front()
	if front == nil || tl.more(v, front.Value) {
		list.PushFront(v)
	} else {
		el := list.Back()
		for tl.more(v, el.Value) {
			el = el.Prev()
		}
		list.InsertAfter(v, el)
	}

	return true
}

func (tl *TopList) Elements() *list.List {
	return tl.list
}

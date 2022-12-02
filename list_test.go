package fun

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestTypes(t *testing.T) {
	suite.Run(t, new(ListSuite))
}

type ListSuite struct {
	suite.Suite
}

func (s *ListSuite) TestQueue() {
	q := NewQueue[string]()
	s.Equal(true, q.IsEmpty())
	s.Equal(0, q.Size())

	q = q.
		Put("hi").
		Put("there")
	s.Equal(false, q.IsEmpty())
	s.Equal(2, q.Size())

	result, success := q.Take()
	s.Equal(true, success)
	s.Equal("hi", result)
	s.Equal(false, q.IsEmpty())
	s.Equal(1, q.Size())

	result, success = q.Take()
	s.Equal(true, success)
	s.Equal("there", result)
	s.Equal(true, q.IsEmpty())
	s.Equal(0, q.Size())

	result, success = q.Take()
	s.Equal(false, success)
	s.Equal("", result)
	s.Equal(true, q.IsEmpty())
	s.Equal(0, q.Size())

	q = q.
		Put("beep").
		Put("boop")
	s.Equal(false, q.IsEmpty())
	s.Equal(2, q.Size())

	result, success = q.Take()
	s.Equal(true, success)
	s.Equal("beep", result)
	s.Equal(false, q.IsEmpty())
	s.Equal(1, q.Size())

	q = q.
		Put("yeehaw")
	s.Equal(false, q.IsEmpty())
	s.Equal(2, q.Size())

	result, success = q.Take()
	s.Equal(true, success)
	s.Equal("boop", result)
	s.Equal(false, q.IsEmpty())
	s.Equal(1, q.Size())

	result, success = q.Take()
	s.Equal(true, success)
	s.Equal("yeehaw", result)
	s.Equal(true, q.IsEmpty())
	s.Equal(0, q.Size())

	result, success = q.Take()
	s.Equal(false, success)
	s.Equal("", result)
	s.Equal(true, q.IsEmpty())
	s.Equal(0, q.Size())

	q = q.
		Put("blah").
		Put("wah").
		Put("shah")
	var results []string
	for data := range q.Iterator() {
		results = append(results, data)
	}
	s.Equal([]string{"blah", "wah", "shah"}, results)
}

func (s *ListSuite) TestStack() {
	c := NewStack[string]()
	s.Equal(true, c.IsEmpty())
	s.Equal(0, c.Size())

	c = c.
		Put("hi").
		Put("there")
	s.Equal(false, c.IsEmpty())
	s.Equal(2, c.Size())

	result, success := c.Take()
	s.Equal(true, success)
	s.Equal("there", result)
	s.Equal(false, c.IsEmpty())
	s.Equal(1, c.Size())

	result, success = c.Take()
	s.Equal(true, success)
	s.Equal("hi", result)
	s.Equal(true, c.IsEmpty())
	s.Equal(0, c.Size())

	result, success = c.Take()
	s.Equal(false, success)
	s.Equal("", result)
	s.Equal(true, c.IsEmpty())
	s.Equal(0, c.Size())

	c = c.
		Put("beep").
		Put("boop")
	s.Equal(false, c.IsEmpty())
	s.Equal(2, c.Size())

	result, success = c.Take()
	s.Equal(true, success)
	s.Equal("boop", result)
	s.Equal(false, c.IsEmpty())
	s.Equal(1, c.Size())

	c = c.
		Put("yeehaw")
	s.Equal(false, c.IsEmpty())
	s.Equal(2, c.Size())

	result, success = c.Take()
	s.Equal(true, success)
	s.Equal("yeehaw", result)
	s.Equal(false, c.IsEmpty())
	s.Equal(1, c.Size())

	result, success = c.Take()
	s.Equal(true, success)
	s.Equal("beep", result)
	s.Equal(true, c.IsEmpty())
	s.Equal(0, c.Size())

	result, success = c.Take()
	s.Equal(false, success)
	s.Equal("", result)
	s.Equal(true, c.IsEmpty())
	s.Equal(0, c.Size())

	c = c.
		Put("blah").
		Put("wah").
		Put("shah")
	var results []string
	for data := range c.Iterator() {
		results = append(results, data)
	}
	s.Equal([]string{"shah", "wah", "blah"}, results)
}

func (s *ListSuite) TestLinkedListPutFront() {
	l := NewLinkedList[string]()
	s.Equal(true, l.IsEmpty())
	s.Equal("", l.Head())
	s.Equal("", l.Tail())

	l = l.PutFront("hi")
	s.Equal(false, l.IsEmpty())
	s.Equal("hi", l.Head())
	s.Equal("hi", l.Tail())
	s.Equal("hi", l.Get(0))
	s.Equal("", l.Get(1))

	l = l.PutFront("there")
	s.Equal(false, l.IsEmpty())
	s.Equal("there", l.Head())
	s.Equal("hi", l.Tail())
	s.Equal("there", l.Get(0))
	s.Equal("hi", l.Get(1))
	s.Equal("", l.Get(2))
}

func (s *ListSuite) TestLinkedListPutBack() {
	l := NewLinkedList[string]()
	s.Equal(true, l.IsEmpty())
	s.Equal("", l.Head())
	s.Equal("", l.Tail())

	l = l.PutBack("hi")
	s.Equal(false, l.IsEmpty())
	s.Equal("hi", l.Head())
	s.Equal("hi", l.Tail())
	s.Equal("hi", l.Get(0))
	s.Equal("", l.Get(1))

	l = l.PutBack("there")
	s.Equal(false, l.IsEmpty())
	s.Equal("hi", l.Head())
	s.Equal("there", l.Tail())
	s.Equal("hi", l.Get(0))
	s.Equal("there", l.Get(1))
	s.Equal("", l.Get(2))
}

func (s *ListSuite) TestLinkedListIterator() {
	l := NewLinkedList[string]().
		PutFront("hi").
		PutBack("there").
		PutFront("billy")

	var results []string
	for data := range l.Iterator() {
		results = append(results, data)
	}
	s.Equal([]string{"billy", "hi", "there"}, results)
}

func (s *ListSuite) TestLinkedListTake() {
	l := NewLinkedList[string]().
		PutFront("hi").
		PutBack("there").
		PutFront("billy")

	data, success := l.TakeFront()
	s.Equal(true, success)
	s.Equal("billy", data)
}

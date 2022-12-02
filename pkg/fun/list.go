package fun

type Node[T any] struct {
	Prev *Node[T]
	Next *Node[T]
	Data T
}

func NewNode[T any](data T) *Node[T] {
	return &Node[T]{
		Data: data,
	}
}

func NewEmptyNode[T any]() *Node[T] {
	return &Node[T]{}
}

func (n *Node[T]) SetPrev(other *Node[T]) *Node[T] {
	n.Prev = other
	return n
}

func (n *Node[T]) SetNext(other *Node[T]) *Node[T] {
	n.Next = other
	return n
}

func (n *Node[T]) PutPrev(data T) *Node[T] {
	prev := NewNode[T](data)
	prev.Next = n
	n.Prev = prev
	return prev
}

func (n *Node[T]) PutNext(data T) *Node[T] {
	next := NewNode[T](data)
	next.Prev = n
	n.Next = next
	return next
}

func (n *Node[T]) TakeNext() (T, *Node[T]) {
	data := n.Data
	next := n.Next
	n.Next = nil
	if next != nil {
		next.Prev = nil
	}
	return data, next
}

func (n *Node[T]) TakePrev() (T, *Node[T]) {
	data := n.Data
	prev := n.Prev
	n.Prev = nil
	if prev != nil {

		prev.Next = nil
	}
	return data, prev
}

type LinkedList[T any] struct {
	Len  int
	head *Node[T]
	tail *Node[T]
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) IsEmpty() bool {
	return l.Len == 0
}

func (l *LinkedList[T]) Head() T {
	if l.head == nil {
		var zero T
		return zero
	} else {
		return l.head.Data
	}
}

func (l *LinkedList[T]) Tail() T {
	if l.tail == nil {
		var zero T
		return zero
	} else {
		return l.tail.Data
	}
}

func (l *LinkedList[T]) Get(i int) T {
	var zero T
	current := l.head
	for k := 0; k < i; k++ {
		if current.Next == nil {
			return zero
		} else {
			current = current.Next
		}
	}
	return current.Data
}

func (l *LinkedList[T]) PutFront(data T) *LinkedList[T] {
	if l.head == nil {
		l.head = NewNode[T](data)
		l.tail = l.head
	} else {
		l.head = l.head.PutPrev(data)
	}
	l.Len += 1
	return l
}

func (l *LinkedList[T]) PutBack(data T) *LinkedList[T] {
	if l.tail == nil {
		l.tail = NewNode[T](data)
		l.head = l.tail
	} else {
		l.tail = l.tail.PutNext(data)
	}
	l.Len += 1
	return l
}

func (l *LinkedList[T]) TakeFront() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	} else {
		var data T
		data, l.head = l.head.TakeNext()
		if l.head == nil {
			l.tail = nil
		}
		l.Len -= 1
		return data, true
	}
}

func (l *LinkedList[T]) TakeBack() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	} else {
		var data T
		data, l.tail = l.tail.TakePrev()
		if l.tail == nil {
			l.head = nil
		}
		l.Len -= 1
		return data, true
	}
}

func (l *LinkedList[T]) Iterator() <-chan T {
	ch := make(chan T, l.Len)
	go func() {
		for current := l.head; current != nil; current = current.Next {
			ch <- current.Data
		}
		close(ch)
	}()
	return ch
}

type Collection[T any] interface {
	Put(data T) Collection[T]
	Take() (T, bool)
	Size() int
	IsEmpty() bool
	Iterator() <-chan T
}

type Queue[T any] struct {
	l *LinkedList[T]
}

func NewQueue[T any]() Collection[T] {
	return &Queue[T]{
		l: NewLinkedList[T](),
	}
}

func (q *Queue[T]) Size() int {
	return q.l.Len
}

func (q *Queue[T]) IsEmpty() bool {
	return q.l.IsEmpty()
}

func (q *Queue[T]) Put(data T) Collection[T] {
	q.l.PutBack(data)
	return q
}

func (q *Queue[T]) Take() (T, bool) {
	if q.l.IsEmpty() {
		var zero T
		return zero, false
	} else {
		data, _ := q.l.TakeFront()
		return data, true
	}
}

func (q *Queue[T]) Iterator() <-chan T {
	return q.l.Iterator()
}

type Stack[T any] struct {
	l *LinkedList[T]
}

func NewStack[T any]() Collection[T] {
	return &Stack[T]{
		l: NewLinkedList[T](),
	}
}

func (s *Stack[T]) Put(data T) Collection[T] {
	s.l.PutFront(data)
	return s
}

func (s *Stack[T]) Take() (T, bool) {
	return s.l.TakeFront()
}

func (s *Stack[T]) Size() int {
	return s.l.Len
}

func (s *Stack[T]) IsEmpty() bool {
	return s.l.IsEmpty()
}

func (s *Stack[T]) Iterator() <-chan T {
	return s.l.Iterator()
}

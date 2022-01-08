package linklist

type LnkNode struct {
	Next *LnkNode
	item interface{}
}

type Lnk struct {
	head    LnkNode
	tail    *LnkNode
	elCount int
}

func (l *Lnk) DoInit() {
	l.tail = &l.head
}

func (l *Lnk) AppendAtTail(item interface{}) {
	l.tail.Next = &LnkNode{
		item: item,
	}
	l.tail = l.tail.Next
	l.elCount++
}

func (l *Lnk) AppendAtFront(item interface{}) {
	oldNext := l.head.Next
	l.head.Next = &LnkNode{
		item: item,
		Next: oldNext,
	}
	if l.tail == &l.head {
		l.tail = l.head.Next
	}
	l.elCount++
}

func (l Lnk) ElementCount() int {
	return l.elCount
}

func (l *Lnk) Extract(tester func(interface{}) bool) interface{} {
	var prev = &l.head
	var start interface{}
	for ptr := l.head.Next; ptr != nil; ptr = ptr.Next {
		if tester(ptr.item) {
			start = ptr.item
			prev.Next = ptr.Next
			if l.tail == ptr {
				l.tail = prev
			}
			l.elCount--
			break
		}
		prev = ptr
	}
	return start
}

func (l *Lnk) ConstForEach(forEach func(interface{})) {
	for ptr := l.head.Next; ptr != nil; ptr = ptr.Next {
		forEach(ptr.item)
	}
}

func NewLnkArray(num int) []Lnk {
	distr := make([]Lnk, num)
	for i := 0; i < len(distr); i++ {
		distr[i].DoInit()
	}
	return distr
}

// Do not return Lnk
func NewLnkHead() *Lnk {
	var rv Lnk
	rv.DoInit()
	return &rv
}

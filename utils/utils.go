package utils

import "git.enflame.cn/hai.bai/dmaster/codec"

type LnkNode struct {
	Next     *LnkNode
	dpfEvent codec.DpfEvent
}

type Lnk struct {
	head    LnkNode
	tail    *LnkNode
	elCount int
}

func (l *Lnk) DoInit() {
	l.tail = &l.head
}

func (l *Lnk) AppendNode(d codec.DpfEvent) {
	l.tail.Next = &LnkNode{
		dpfEvent: d,
	}
	l.tail = l.tail.Next
	l.elCount++
}

func (l *Lnk) Extract(target codec.DpfEvent) *codec.DpfEvent {
	var prev = &l.head
	var start *codec.DpfEvent
	for ptr := l.head.Next; ptr != nil; ptr = ptr.Next {
		if ptr.dpfEvent.PacketID+1 == target.PacketID {
			start = &ptr.dpfEvent
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

func NewLnkArray(num int) []Lnk {
	distr := make([]Lnk, num)
	for i := 0; i < len(distr); i++ {
		distr[i].DoInit()
	}
	return distr
}

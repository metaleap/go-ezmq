package ezmq

import (
	"time"
)

//	Just a minimal demo "line-of-business object" type that ezmq can pub/sub.
//	This type as well as BizFoo showcase just how easily one would extend one's
//	in-house / company-specific ezMQ wrapper library by additional custom
//	message types as desired.
type BizEvent struct {
	Id   string
	Name string
	Date time.Time
}

func NewBizEvent(id string, name string) *BizEvent {
	return NewBizEventAt(id, name, time.Now())
}

func NewBizEventAt(id string, name string, date time.Time) *BizEvent {
	return &BizEvent{Id: id, Name: name, Date: date}
}

//	A well-typed (to `BizEvent`) wrapper around `Queue.SubscribeTo`.
func (q *Queue) SubscribeToBizEvents(subscribers ...func(*BizEvent)) (err error) {
	makeEvent := func() interface{} { return BizEvent{} }
	return q.SubscribeTo(makeEvent, func(ptr interface{}) {
		evt, _ := ptr.(*BizEvent)
		for _, onevent := range subscribers {
			onevent(evt)
		}
	})
}

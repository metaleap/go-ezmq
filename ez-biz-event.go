package ezmq

import (
	"time"
)

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

func (ex *Exchange) PublishBizEvent(evt *BizEvent) error {
	return ex.Publish(evt)
}

func (q *Queue) PublishBizEvent(evt *BizEvent) error {
	return q.Publish(evt)
}

func (q *Queue) SubscribeToBizEvents(subscribers ...func(*BizEvent)) (err error) {
	makeEmptyBizEventForDeserialization := func() interface{} { return BizEvent{} }
	return q.SubscribeTo(makeEmptyBizEventForDeserialization, func(ptr interface{}) {
		evt, ok := ptr.(*BizEvent)
		if ok && evt != nil { // yeah tautological but explicit, we want a non-nil whatever the semantics of "ok"
			for _, onevent := range subscribers {
				onevent(evt)
			}
		} // else it was something other than an BizEvent arriving in this queue/channel/routing-key/yadda-yadda and we're not listening to that, we're subscribed to BizEvents here
	})
}

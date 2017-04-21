package ezmq

import (
	"time"
)

//	Just a minimal demo "line-of-business object" type that ezmq can pub/sub.
//	This type as well as BizFoo showcase just how easily one would extend one's
//	in-house / company-specific ezMQ wrapper library by additional custom
//	message types as desired. In fact, both are ripe candidates for codegen-ing
//	the methods associated with one's line-of-business struct:
//	`Exchange.PublishXyz`, `Queue.PublishXyz`, and `Queue.SubscribeToXyzs`.
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

//	Convenience short-hand for `ex.Publish(evt)`
func (ex *Exchange) PublishBizEvent(evt *BizEvent) error {
	return ex.Publish(evt)
}

//	Convenience short-hand for `q.Publish(evt)`
func (q *Queue) PublishBizEvent(evt *BizEvent) error {
	return q.Publish(evt)
}

//	A well-typed (to `BizEvent`) wrapper around `Queue.SubscribeTo`.
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

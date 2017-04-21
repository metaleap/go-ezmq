package ezmq

import (
	"time"
)

type Event struct {
	Id   string
	Name string
	Date time.Time
}

func NewEvent(id string, name string) *Event {
	return NewEventAt(id, name, time.Now())
}

func NewEventAt(id string, name string, date time.Time) *Event {
	return &Event{Id: id, Name: name, Date: date}
}

func (ex *Exchange) PublishEvent(evt *Event) error {
	return ex.publish(evt)
}

func (q *Queue) PublishEvent(evt *Event) error {
	return q.publish(evt)
}

func (q *Queue) SubscribeToEvents(subscribers ...func(*Event)) (err error) {
	makeEmptyEventForDeserialization := func() interface{} { return Event{} }
	return q.SubscribeTo(makeEmptyEventForDeserialization, func(evt interface{}) {
		event, ok := evt.(*Event)
		if ok && event != nil { // yeah tautological but explicit, we want a non-nil whatever the semantics of "ok"
			for _, onevent := range subscribers {
				onevent(event)
			}
		} // else it was something other than an Event arriving in this queue/channel/routing-key/yadda-yadda and we're not listening to that, we're subscribed to Events here
	})
}

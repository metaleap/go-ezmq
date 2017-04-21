package ezmq

type Foo struct {
	Bar bool
	Baz int
}

func NewFoo(bar bool, baz int) *Foo {
	return &Foo{Bar: bar, Baz: baz}
}

func (ex *Exchange) PublishFoo(foo *Foo) error {
	return ex.Publish(foo)
}

func (q *Queue) PublishFoo(foo *Foo) error {
	return q.Publish(foo)
}

func (q *Queue) SubscribeToFoos(subscribers ...func(*Foo)) (err error) {
	makeEmptyFooForDeserialization := func() interface{} { return Foo{} }
	return q.SubscribeTo(makeEmptyFooForDeserialization, func(evt interface{}) {
		foo, ok := evt.(*Foo)
		if ok && foo != nil { // yeah tautological but explicit, we want a non-nil whatever the semantics of "ok"
			for _, onfoo := range subscribers {
				onfoo(foo)
			}
		} // else it was something other than an Foo arriving in this queue/channel/routing-key/yadda-yadda and we're not listening to that, we're subscribed to Foos here
	})
}

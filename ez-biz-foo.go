package ezmq

type BizFoo struct {
	Bar bool
	Baz int
}

func NewBizFoo(bar bool, baz int) *BizFoo {
	return &BizFoo{Bar: bar, Baz: baz}
}

func (ex *Exchange) PublishBizFoo(foo *BizFoo) error {
	return ex.Publish(foo)
}

func (q *Queue) PublishBizFoo(foo *BizFoo) error {
	return q.Publish(foo)
}

func (q *Queue) SubscribeToBizFoos(subscribers ...func(*BizFoo)) (err error) {
	makeEmptyBizFooForDeserialization := func() interface{} { return BizFoo{} }
	return q.SubscribeTo(makeEmptyBizFooForDeserialization, func(evt interface{}) {
		foo, ok := evt.(*BizFoo)
		if ok && foo != nil { // yeah tautological but explicit, we want a non-nil whatever the semantics of "ok"
			for _, onfoo := range subscribers {
				onfoo(foo)
			}
		} // else it was something other than an BizFoo arriving in this queue/channel/routing-key/yadda-yadda and we're not listening to that, we're subscribed to BizFoos here
	})
}

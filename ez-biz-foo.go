package ezmq

//	Just a minimal demo "line-of-business object" type that ezmq can pub/sub.
//	This type as well as BizEvent showcase just how easily one would extend one's
//	in-house / company-specific ezMQ wrapper library by additional custom
//	message types as desired. In fact, both are ripe candidates for codegen-ing
//	the well-typed wrapper methods associated with one's line-of-business struct:
//	`Exchange.PublishXyz`, `Queue.PublishXyz`, and `Queue.SubscribeToXyzs`.
type BizFoo struct {
	Bar bool
	Baz int
}

func NewBizFoo(bar bool, baz int) *BizFoo {
	return &BizFoo{Bar: bar, Baz: baz}
}

//	Convenience short-hand for `ex.Publish(foo)`
func (ex *Exchange) PublishBizFoo(foo *BizFoo) error {
	return ex.Publish(foo)
}

//	Convenience short-hand for `q.Publish(foo)`
func (q *Queue) PublishBizFoo(foo *BizFoo) error {
	return q.Publish(foo)
}

//	A well-typed (to `BizFoo`) wrapper around `Queue.SubscribeTo`.
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

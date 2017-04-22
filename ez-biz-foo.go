package ezmq

//	Just a minimal demo "line-of-business object" type that ezmq can pub/sub.
//	This type as well as BizEvent showcase just how easily one would extend one's
//	in-house / company-specific ezMQ wrapper library by additional custom
//	message types as desired.
type BizFoo struct {
	Bar bool
	Baz int
}

//	A well-typed (to `BizFoo`) wrapper around `Queue.SubscribeTo`.
func (q *Queue) SubscribeToBizFoos(subscribers ...func(*BizFoo)) (err error) {
	makeFoo := func() interface{} { return BizFoo{} }
	return q.SubscribeTo(makeFoo, func(ptr interface{}) {
		foo, _ := ptr.(*BizFoo)
		for _, onfoo := range subscribers {
			onfoo(foo)
		}
	})
}

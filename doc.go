//	Provides a higher-level, type-driven message-queuing API wrapping RabbitMQ / amqp.
//
//	## Scenarios
//
//	"Line-of-business object" types used here, `BizEvent` and `BizFoo`, are included for
//	*demo* purposes, to showcase how easily one may "type-safe-ish"ly broadcast and subscribe-to
//	any kind of custom, in-house struct type; ezmq employs JSON serialization only for now.
//
//	Pseudo-code ignores all the `error`s returned that it should in real life check:
//
//	### Simple publishing via Queue:
//
//	    ctx := ezmq.LocalCtx                                        // guest:guest@localhost:5672
//	    defer ctx.Close()
//	    var qcfg *ezmq.QueueConfig = nil                            // nil = use 'prudent' defaults
//	    //
//	    qe := ctx.Queue('myevents', qcfg)
//	    qe.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))
//	    qf := ctx.Queue('myfoos', qcfg)
//	    qf.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//	    // some more for good measure:
//	    qe.Publish(ezmq.NewBizEvent("evt2", "DatEvent"))
//	    qf.Publish(&ezmq.BizFoo{ Baz: 20 })                         // same thing just untyped
//	    qe.Publish(ezmq.NewBizEvent("evt3", "SomeEvent"))           // ditto
//
//	### Simple subscribing via Queue:
//
//	    onBizEvent := func(evt *ezmq.BizEvent) {
//	        println(evt.Name)
//	    }
//	    qe.SubscribeToBizEvents(onBizEvent)
//	    qf.SubscribeToBizFoos(func(foo *ezmq.Foo) { mylogger.LogAnything(foo) })
//	    for true { /* we loop until we won't */ }
//
//	### Multiple subscribers via Exchange:
//
//	    qm := ctx.Queue('', qcfg)   //  name MUST be empty
//	    var xcfg *ezmq.ExchangeConfig = nil // as usual, nil = defaults
//	    ex := ctx.Exchange('mybroadcast', xcfg, qm)  //  only pass `Queue`s that were declared with empty `name`
//	    ex.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))  //  publish via `Exchange`, not via `Queue`, same API
//	    ex.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//	    ex.Publish(ezmq.NewBizEvent("evt2", "DatEvent")) // same thing just untyped
//	    ex.Publish(&ezmq.BizFoo{ Baz: 20 }) // ditto
//
//	### Enabling multiple worker instances:
//
//	    // Pass this then to ctx.Queue()
//	    var qcfg *ezmq.QueueConfig = ezmq.ConfigDefaultsQueue
//	    qcfg.Pub.Persistent = true
//	    qcfg.Pub.QosMultipleWorkerInstances = true
//
//	    // Pass this then to ctx.Exchange(), if one is used
//	    var xcfg *ezmq.ExchangeConfig = ezmq.ConfigDefaultsExchange
//	    xcfg.Pub.Persistent = true
//	    xcfg.Pub.QosMultipleWorkerInstances = true
//	    // Rest as usual
//
//
package ezmq

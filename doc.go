//	A higher-level, type-driven, simplified message-queuing API wrapping+hiding+streamlining RabbitMQ & streadway/amqp under the hood (while retaining much of their flexibility & powers)
//
//	## High-Level API Workflow:
//
//	* make a `Context` (later, when done, `Close` it)
//	* for **simple messaging**, use it to declare a named `Queue`, then
//	    * `Queue.Publish(anything)`
//	    * `Queue.SubscribeTo<Thing>s(myOnThingHandlers...)`
//	* to publish to **multiple subscribers**
//	    * use the `Context` and an unnamed `Queue` to declare an `Exchange`,
//	    * then `Exchange.Publish(anything)`
//	* for multiple **worker instances**, set 2 `bool`s, as below
//
//	## Example Scenarios
//
//	"Line-of-business object" types used here, `BizEvent` and `BizFoo`, are included for
//	*demo* purposes, to showcase how easily one may "type-safe-ish"ly broadcast and subscribe-to
//	any kind of custom, in-house struct type.
//
//	Pseudo-code ignores all the `error`s returned that it should in real life check:
//
//	### Simple publishing via Queue:
//
//	    ctx := ezmq.LocalCtx                                        // guest:guest@localhost:5672
//	    defer ctx.Close()
//	    var qcfg *ezmq.QueueConfig = nil                            // nil = use 'prudent' defaults
//
//	    qe,_ := ctx.Queue('myevents', qcfg)
//	    qe.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))
//
//	    qf,_ := ctx.Queue('myfoos', qcfg)
//	    qf.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//
//	    // some more for good measure:
//	    qe.Publish(ezmq.NewBizEvent("evt2", "DatEvent"))
//	    qf.Publish(&ezmq.BizFoo{ Baz: 20 })
//	    qe.Publish(ezmq.NewBizEvent("evt3", "SomeEvent"))
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
//	    qm,_ := ctx.Queue('', qcfg)   //  name MUST be empty
//	    var xcfg *ezmq.ExchangeConfig = nil // as above, nil = defaults
//	    ex,_ := ctx.Exchange('mybroadcast', xcfg, qm)  //  only pass `Queue`s that were declared with empty `name`
//
//	    ex.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))  //  publish via `Exchange`, not via `Queue`, same API
//	    ex.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//	    ex.Publish(ezmq.NewBizEvent("evt2", "DatEvent"))
//	    ex.Publish(&ezmq.BizFoo{ Baz: 20 })
//
//	### Enabling multiple worker instances:
//
//	    // Pass this then to ctx.Queue()
//	    var qcfg *ezmq.QueueConfig = ezmq.ConfigDefaultsQueue
//	    qcfg.Pub.Persistent = true
//	    qcfg.QosMultipleWorkerInstances = true
//
//	    // Pass this then to ctx.Exchange(), if one is used (WITH a queue declared with the above)
//	    var xcfg *ezmq.ExchangeConfig = ezmq.ConfigDefaultsExchange
//	    xcfg.Pub.Persistent = true
//
//	    // Rest as usual
//
//
package ezmq

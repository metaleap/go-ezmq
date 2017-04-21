//	Provides a higher-level, type-driven message-queuing API wrapping RabbitMQ / amqp.
//
//	## Scenarios
//
//	Pseudo-code ignores all the `error`s returned that it should in real life check:
//
//	### Simple publishing via Queue:
//
//	    ctx := ezmq.LocalCtx // guest:guest@localhost:5672
//	    defer ctx.Close()
//	    var qcfg *ezmq.QueueConfig = nil // that's OK
//
//	    qe := ctx.Queue('myevents', qcfg)
//	    qe.PublishBizEvent(ezmq.NewBizEvent("evt1", "DisEvent"))
//	    qf := ctx.Queue('myfoos', qcfg)
//	    qf.PublishBizFoo(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//	    qe.PublishBizEvent(ezmq.NewBizEvent("evt2", "DatEvent"))
//	    qf.Publish(&ezmq.BizFoo{ Baz: 20 }) // same thing just untyped
//	    qe.Publish(ezmq.NewBizEvent("evt3", "SomeEvent")) // ditto
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
//	    qm := ctx.Queue('', qcfg)
//	    var xcfg *ezmq.ExchangeConfig = nil // that's OK
//	    ex := ctx.Exchange('mybroadcast', xcfg, qm)
//	    ex.PublishBizEvent(ezmq.NewBizEvent("evt1", "DisEvent"))
//	    ex.PublishBizFoo(&ezmq.BizFoo{ Bar: true, Baz: 10 })
//	    ex.Publish(ezmq.NewBizEvent("evt2", "DatEvent")) // same thing just untyped
//	    ex.Publish(&ezmq.BizFoo{ Baz: 20 }) // ditto
//
//	### Enabling multiple worker instances:
//
//	    // Prior to ctx.Queue()
//	    var qcfg *ezmq.QueueConfig = ezmq.ConfigDefaultsQueue
//	    qcfg.Pub.Persistent = true
//	    qcfg.Pub.QosMultipleWorkerInstances = true
//
//	    // If used, prior to ctx.Exchange()
//	    var xcfg *ezmq.ExchangeConfig = ezmq.ConfigDefaultsExchange
//	    xcfg.Pub.Persistent = true
//	    xcfg.Pub.QosMultipleWorkerInstances = true
//
//	    // Rest as above
package ezmq

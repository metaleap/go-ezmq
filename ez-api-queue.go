package ezmq

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

//	Used to publish and to subscribe. ONLY to be constructed via
//	`Context.Queue()`, and fields are not to be mutated thereafter! It remains
//	associated with that `Context` for all its `Publish` and `SubscribeTo` calls.
type Queue struct {
	//	Contains either the `name` given at initialization to `Context.Queue()`,
	//	or a random identifier if that was empty: in such cases, this `Queue`
	//	MUST be used to bind to an `Exchange` constructed via `Context.Exchange()`,
	//	and the `Config` will be set to a newly created copy of the effective
	//	`cfg` in `Context.Queue()`, with tweaked `Durable` and `Exclusive` fields.
	Name string

	//	Set to sensible defaults of `ConfigDefaultsQueue` at initialization.
	Config *QueueConfig

	ctx *Context
}

//	Specialist tweaks for declaring a `Queue` to the backing message-queue.
type QueueConfig struct {
	Durable                    bool
	AutoDelete                 bool
	Exclusive                  bool
	NoWait                     bool
	QosMultipleWorkerInstances bool
	Pub                        TweakPub
	Sub                        TweakSub
	Args                       map[string]interface{}
}

var (
	//	Can be modified, but mustn't be `nil`. Initially contains prudent
	//	defaults quite sensible during prototyping, until you *know* what few
	//	things you need to tweak and why.
	//	Used by `Context.Queue()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsQueue = &QueueConfig{Durable: true, Sub: TweakSub{AutoAck: true}}
)

//	Declares a queue with the specified `name` for publishing and subscribing.
//	If `cfg` is `nil`, the current `ConfigDefaultsQueue` is used. For `name`,
//	DO refer to the docs on `Queue.Name`.
func (ctx *Context) Queue(name string, cfg *QueueConfig) (q *Queue, err error) {
	if cfg == nil {
		cfg = ConfigDefaultsQueue
	}
	q = &Queue{ctx: ctx, Config: cfg}
	if err = ctx.ensureConnectionAndChannel(); err == nil {
		exclusive, durable := cfg.Exclusive, cfg.Durable
		if len(name) == 0 {
			exclusive = true
			durable = false
		}
		if exclusive != cfg.Exclusive || durable != cfg.Durable {
			nucfg := *cfg
			nucfg.Exclusive, nucfg.Durable = exclusive, durable
			q.Config = &nucfg
			cfg = q.Config
		}
		var _q amqp.Queue
		if _q, err = ctx.ch.QueueDeclare(name, durable, cfg.AutoDelete, exclusive, cfg.NoWait, cfg.Args); err == nil {
			q.Name = _q.Name // if `name` was empty, this will be a random-name
			if cfg.QosMultipleWorkerInstances {
				err = ctx.ch.Qos(1, 0, false)
			}
		}
	}
	if err != nil {
		q = nil
	}
	return
}

//	Serializes the specified `obj` to JSON and publishes it to this exchange.
func (q *Queue) Publish(obj interface{}) error {
	return q.ctx.publish(obj, "", q.Name, &q.Config.Pub)
}

//	Generic subscription mechanism used by the more convenient well-typed
//	wrapper functions such as `SubscribeToEvents` and `SubscribeToFoos`:
//
//	Subscribe to messages only of the Type returned by the specified
//	`makeEmptyObjForDeserialization` constructor function used to allocate a
//	new empty/zeroed non-pointer struct value whenever attempting to deserialize
//	a message received from this `Queue`. If that succeeds, a pointer to that
//	value is passed to the specified `onMsg` event handler: this will always be
//	passed a non-nil pointer to the value (now populated with data) returned by
//	`makeEmptyObjForDeserialization`, therefore safe to cast back to the Type.
func (q *Queue) SubscribeTo(makeEmptyObjForDeserialization func() interface{}, onMsg func(interface{})) (err error) {
	if err = q.ctx.ensureConnectionAndChannel(); err == nil {
		cfgSub := q.Config.Sub
		var msgchan <-chan amqp.Delivery
		msgchan, err = q.ctx.ch.Consume(q.Name, cfgSub.Consumer, cfgSub.AutoAck && !q.Config.QosMultipleWorkerInstances, q.Config.Exclusive, cfgSub.NoLocal, q.Config.NoWait, q.Config.Args)
		if err == nil && msgchan != nil {
			listen := func() {
				for msgdelivery := range msgchan {
					if !cfgSub.AutoAck {
						if ackerr := msgdelivery.Ack(false); ackerr != nil {
							panic("ezmq.Queue.SubscribeTo$listen$msgdelivery.Ack: failed to ack, but error logging wasn't ordered, time to implement!")
						}
					}
					obj := makeEmptyObjForDeserialization()
					jsonerr := json.Unmarshal(msgdelivery.Body, &obj)
					var ptr interface{}
					if jsonerr == nil { // else some sort of msg passed by not fitting into obj and thus of no interest to *these* here subscribers
						ptr = &obj
					}
					if ptr != nil {
						onMsg(ptr)
					}
				}
			}
			go listen()
		}
	}
	return
}

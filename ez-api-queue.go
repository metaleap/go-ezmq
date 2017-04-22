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
	Config QueueConfig

	ctx *Context
}

//	Specialist tweaks for declaring a `Queue` to the backing message-queue.
//	If you don't know their meaning, you're best off keeping our defaults until admins/dev-ops decide otherwise.
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
	//	Can be modified. Initially contains prudent defaults quite sensible
	//	during prototyping, until you *know* what few things you need to tweak and why.
	//	Used by `Context.Queue()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsQueue = QueueConfig{Durable: true, Sub: TweakSub{AutoAck: true}}
)

//	Declares a queue with the specified `name` for publishing and subscribing.
//	If `cfg` is `nil`, a copy of the current `ConfigDefaultsQueue` is used
//	for `q.Config`, else a copy of `cfg`. For `name`, DO refer to `Queue.Name`.
func (ctx *Context) Queue(name string, cfg *QueueConfig) (q *Queue, err error) {
	q = &Queue{ctx: ctx, Config: ConfigDefaultsQueue}
	if cfg != nil {
		q.Config = *cfg
	}
	if isforbindingtoexchange := len(name) == 0; isforbindingtoexchange {
		q.Config.Durable, q.Config.Exclusive = false, true
	}
	if err = ctx.ensureConnectionAndChannel(); err == nil {
		var queueserverstate amqp.Queue
		if queueserverstate, err = ctx.ch.QueueDeclare(name, q.Config.Durable, q.Config.AutoDelete, q.Config.Exclusive, q.Config.NoWait, q.Config.Args); err == nil {
			q.Name = queueserverstate.Name // if `name` was empty, this will be a random-name
			if q.Config.QosMultipleWorkerInstances {
				err = ctx.ch.Qos(1, 0, false)
			}
		}
	}
	if err != nil {
		q = nil
	}
	return
}

//	Serializes the specified `obj` and publishes it to this exchange.
func (q *Queue) Publish(obj interface{}) error {
	return q.ctx.publish(obj, "", q.Name, &q.Config.Pub)
}

//	Generic subscription mechanism used by the more convenient well-typed
//	wrapper functions such as `SubscribeToBizEvents` and `SubscribeToBizFoos`:
//
//	Subscribe to messages only of the Type returned by the specified
//	`makeEmptyObjForDeserialization` constructor function used to allocate a
//	new empty/zeroed non-pointer struct value whenever attempting to deserialize
//	a message received from this `Queue`. If that succeeds, a pointer to that
//	value is passed to the specified `onMsg` event handler: this will always be
//	passed a non-nil pointer to the value (now populated with data) returned by
//	`makeEmptyObjForDeserialization`, therefore safe to cast back to the Type.
func (q *Queue) SubscribeTo(makeEmptyObjForDeserialization func() interface{}, onMsg func(interface{})) (err error) {
	if err = q.ctx.ensureConnectionAndChannel(); err != nil {
		return
	}
	cfgsub := q.Config.Sub
	autoack := cfgsub.AutoAck && !q.Config.QosMultipleWorkerInstances
	var msgchan <-chan amqp.Delivery
	if msgchan, err = q.ctx.ch.Consume(q.Name, cfgsub.Consumer, autoack, q.Config.Exclusive, cfgsub.NoLocal, q.Config.NoWait, q.Config.Args); err != nil || msgchan == nil {
		return
	}
	keepListeningInABackgroundLoopGoroutine := func() {
		for msgdelivery := range msgchan {
			ptr, ackerr := decodeMsgForSubscribers(&msgdelivery, !autoack, makeEmptyObjForDeserialization)
			if ackerr != nil && cfgsub.OnAckError != nil { // we're looping in a goroutine so this one, the client can handle or ignore as preferred
				if keepgoing := cfgsub.OnAckError(ackerr); !keepgoing {
					return
				}
			}
			if ptr != nil {
				onMsg(ptr)
			}
		}
	}
	go keepListeningInABackgroundLoopGoroutine()
	return
}

func decodeMsgForSubscribers(msgDelivery *amqp.Delivery, shouldAck bool, constructor func() interface{}) (ptr interface{}, ackErr error) {
	newvalue := constructor()
	if jsonerr := json.Unmarshal(msgDelivery.Body, &newvalue); jsonerr == nil {
		ptr = &newvalue
	} // else some sort of msg passed by not fitting into newvalue and thus of no interest to *these* here subscribers
	if shouldAck { // TODO: for production, gonna have to refine to full ack semantics (ack/reject/nack) with respect to exclusive
		ackErr = msgDelivery.Ack(false)
	}
	return
}

package ezmq

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Queue struct {
	Config *ConfigQueue
	Name   string

	ctx *Context
}

type ConfigQueue struct {
	Durable                    bool
	AutoDelete                 bool
	Exclusive                  bool
	NoWait                     bool
	Args                       map[string]interface{}
	Pub                        *ConfigPub
	Sub                        *ConfigSub
	QosMultipleWorkerInstances bool
}

var (
	ConfigDefaultsQueue = &ConfigQueue{Durable: true}
)

func (ctx *Context) Queue(name string, cfg *ConfigQueue) (q *Queue, err error) {
	if cfg == nil {
		cfg = ConfigDefaultsQueue // nil implies zeroed-defaults, seems acceptable to amqp
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

func (q *Queue) Publish(obj interface{}) error {
	return q.ctx.publish(obj, "", q.Name, q.Config.Pub)
}

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

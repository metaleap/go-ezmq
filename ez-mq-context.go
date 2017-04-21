//	Provides a higher-level, type-driven message-queuing API wrapping RabbitMQ / amqp.
//
//	## markdown test
package ezmq

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

//	Provides access to the backing message-queue implementation, encapsulating
//	the underlying connection/channel primitives. Set the fields
//	(as fits the project context / local setup) before declaring the `Queue`s or
//	`Exchange`s needed to publish and subscribe, as those calls will connect if
//	the `Context` isn't already connected. Subsequent field mutations are of
//	course ignored as the connection is kept alive. For clean-up or manual /
//	pooled connection strategies, `Context` provides the `Close` method.
type Context struct {
	UserName string
	Password string
	Host     string
	Port     uint16

	conn *amqp.Connection
	ch   *amqp.Channel
}

//	A convenient `Context` for local-machine based prototyping/testing.
var LocalCtx = &Context{UserName: "guest", Password: "guest", Host: "localhost", Port: 5672}

//	Specialist tweaks for `Publish`ing via a `Queue` or an `Exchange`.
type TweakPub struct {
	Mandatory  bool
	Immediate  bool
	Persistent bool
}

//	Specialist tweaks used from within `Queue.SubscribeTo`.
type TweakSub struct {
	Consumer string
	AutoAck  bool
	NoLocal  bool
}

//	Be SURE to call this when done with ezmq, to cleanly dispose of resources.
func (ctx *Context) Close() (chanCloseErr, connCloseErr error) {
	if ctx.ch != nil {
		chanCloseErr = ctx.ch.Close()
		ctx.ch = nil
	}
	if ctx.conn != nil {
		connCloseErr = ctx.conn.Close()
		ctx.conn = nil
	}
	return
}

func (ctx *Context) connectionUri() (uri string) {
	//	there are more efficient ways to concat strings but this won't be called frequently/repeatedly, so we go for readability
	uri = "amqp://"
	if len(ctx.UserName) > 0 {
		uri += ctx.UserName
		if len(ctx.Password) > 0 {
			uri += ":" + ctx.Password
		}
		uri += "@"
	}
	uri += ctx.Host
	if ctx.Port > 0 {
		uri += fmt.Sprintf(":%d", ctx.Port)
	}
	return
}

func (ctx *Context) ensureConnectionAndChannel() (err error) {
	if ctx.conn == nil {
		ctx.conn, err = amqp.Dial(ctx.connectionUri())
	}
	if ctx.conn != nil && ctx.ch == nil {
		ctx.ch, err = ctx.conn.Channel()
	}
	return
}

func (ctx *Context) publish(obj interface{}, exchangeName string, routingKey string, cfgPub *TweakPub) (err error) {
	var msgraw []byte
	if err = ctx.ensureConnectionAndChannel(); err == nil {
		if msgraw, err = json.Marshal(obj); err == nil {
			msgpub := amqp.Publishing{ContentType: "text/plain", Body: msgraw}
			if cfgPub.Persistent {
				msgpub.DeliveryMode = amqp.Persistent
			}
			err = ctx.ch.Publish(exchangeName, routingKey, cfgPub.Mandatory, cfgPub.Immediate, msgpub)
		}
	}
	return
}

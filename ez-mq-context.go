package ezmq

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

type ConfigPub struct {
	Mandatory  bool
	Immediate  bool
	Persistent bool
}

type ConfigSub struct {
	Consumer string
	AutoAck  bool
	NoLocal  bool
}

var (
	ConfigDefaultsPub = &ConfigPub{}
	ConfigDefaultsSub = &ConfigSub{AutoAck: true}
)

var LocalCtx = &Context{User: "guest", Password: "guest", Host: "localhost", Port: 5672}

type Context struct {
	User     string
	Password string
	Host     string
	Port     uint16

	conn *amqp.Connection
	ch   *amqp.Channel
}

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
	if len(ctx.User) > 0 {
		uri += ctx.User
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

func (ctx *Context) publish(obj interface{}, exchangeName string, routingKey string, cfgPub *ConfigPub) (err error) {
	var msgraw []byte
	if err = ctx.ensureConnectionAndChannel(); err == nil {
		if msgraw, err = json.Marshal(obj); err == nil {
			if cfgPub == nil {
				cfgPub = ConfigDefaultsPub
			}
			msgpub := amqp.Publishing{ContentType: "text/plain", Body: msgraw}
			if cfgPub.Persistent {
				msgpub.DeliveryMode = amqp.Persistent
			}
			err = ctx.ch.Publish(exchangeName, routingKey, cfgPub.Mandatory, cfgPub.Immediate, msgpub)
		}
	}
	return
}

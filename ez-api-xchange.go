package ezmq

type Exchange struct {
	Config *ConfigExchange
	Name   string

	ctx *Context
}

type ConfigExchange struct {
	Type       string // fanout/direct/topic/headers --- how about enums one day, go..
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
	Pub        *ConfigPub
	Sub        *ConfigSub
	QueueBind  *ConfigExchangeQueueBind
}

type ConfigExchangeQueueBind struct {
	NoWait     bool
	Args       map[string]interface{}
	RoutingKey string
}

var (
	//	Mustn't be `nil`. Sensible defaults during prototyping, until you *know* what you need to tweak and why.
	//	Used by `Context.Exchange()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsExchange = &ConfigExchange{Durable: true, Type: "fanout"}

	ConfigDefaultsExchangeQueueBind = &ConfigExchangeQueueBind{}
)

func (ctx *Context) Exchange(name string, cfg *ConfigExchange, bindTo *Queue) (ex *Exchange, err error) {
	if cfg == nil {
		cfg = ConfigDefaultsExchange // nil implies zeroed-defaults, seems acceptable to amqp
	}
	ex = &Exchange{ctx: ctx, Name: name, Config: cfg}
	if err = ctx.ensureConnectionAndChannel(); err == nil {
		if err = ctx.ch.ExchangeDeclare(name, cfg.Type, cfg.Durable, cfg.AutoDelete, cfg.Internal, cfg.NoWait, cfg.Args); err == nil {
			err = ctx.ch.QueueBind(bindTo.Name, cfg.QueueBind.RoutingKey, name, cfg.QueueBind.NoWait, cfg.QueueBind.Args)
		}
	}
	if err != nil {
		ex = nil
	}
	return
}

func (ex *Exchange) Publish(obj interface{}) error {
	return ex.ctx.publish(obj, ex.Name, ex.Config.QueueBind.RoutingKey, ex.Config.Pub)
}

package ezmq

//	Used in-place-of / in-conjunction-with a `Queue` when needing to publish to
//	multiple subscribers. ONLY to be constructed via `Context.Exchange()`, and
//	fields are not to be mutated thereafter! It remains associated with that
//	`Context` for all its `Publish` calls.
type Exchange struct {
	//	Shouldn't be empty, although it's legal: that implies the backing
	//	message-queue's "default" exchange --- however there is typically no
	//	need to set up an `Exchange` to use that one, as it's used by default
	//	for example when `Publish`ing directly via a `Queue`.
	Name string

	//	Set to sensible defaults of `ConfigDefaultsExchange` at initialization.
	Config *ExchangeConfig

	ctx *Context
}

//	Specialist tweaks for declaring an `Exchange` to the backing message-queue.
type ExchangeConfig struct {
	Type       string // fanout/direct/topic/headers
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
	Pub        TweakPub
	QueueBind  struct {
		NoWait     bool
		Args       map[string]interface{}
		RoutingKey string
	}
}

var (
	//	Can be modified, but mustn't be `nil`. Initially contains prudent
	//	defaults quite sensible during prototyping, until you *know* what few
	//	things you need to tweak and why.
	//	Used by `Context.Exchange()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsExchange = &ExchangeConfig{Durable: true, Type: "fanout"}
)

//	Declares an "exchange" for publishing to multiple subscribers via the
//	specified `Queue`. (If multiple-subscribers need not be supported, then no
//	need for an `Exchange`: just use a simple `Queue` only.) If `cfg` is `nil`,
//	the current `ConfigDefaultsExchange` is used. For `name`, see `Exchange.Name`.
func (ctx *Context) Exchange(name string, cfg *ExchangeConfig, bindTo *Queue) (ex *Exchange, err error) {
	if cfg == nil {
		cfg = ConfigDefaultsExchange
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

//	Serializes the specified `obj` to JSON and publishes it to this exchange.
func (ex *Exchange) Publish(obj interface{}) error {
	return ex.ctx.publish(obj, ex.Name, ex.Config.QueueBind.RoutingKey, &ex.Config.Pub)
}

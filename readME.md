# ezmq
--
    import "github.com/metaleap/go-ezmq"

Provides a higher-level, type-driven message-queuing API wrapping RabbitMQ /
amqp.

## Scenarios

### Simple publishing via Queue:

    ctx := ezmq.LocalCtx
    defer ctx.Close()
    var cfg *ezmq.QueueConfig = nil // that's OK
    qe := ctx.Queue('myevents', cfg)
    qe.PublishBizEvent(ezmq.NewBizEvent("evt1", "DisEvent"))
    qf := ctx.Queue('myfoos', cfg)
    qf.PublishBizFoo(&ezmq.BizFoo{ Bar: true, Baz: 10 })
    qe.PublishBizEvent(ezmq.NewBizEvent("evt2", "DatEvent"))
    qf.PublishBizFoo(&ezmq.BizFoo{ Baz: 20 })
    qe.PublishBizEvent(ezmq.NewBizEvent("evt3", "SomeEvent"))

## Usage

```go
var (
	//	Can be modified, but mustn't be `nil`. Initially contains prudent
	//	defaults quite sensible during prototyping, until you *know* what few
	//	things you need to tweak and why.
	//	Used by `Context.Exchange()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsExchange = &ExchangeConfig{Durable: true, Type: "fanout"}
)
```

```go
var (
	//	Can be modified, but mustn't be `nil`. Initially contains prudent
	//	defaults quite sensible during prototyping, until you *know* what few
	//	things you need to tweak and why.
	//	Used by `Context.Queue()` if it is passed `nil` for its `cfg` arg.
	ConfigDefaultsQueue = &QueueConfig{Durable: true, Sub: TweakSub{AutoAck: true}}
)
```

```go
var (
	//	A convenient `Context` for local-machine based prototyping/testing.
	LocalCtx = &Context{UserName: "guest", Password: "guest", Host: "localhost", Port: 5672}
)
```

#### type BizEvent

```go
type BizEvent struct {
	Id   string
	Name string
	Date time.Time
}
```

Just a minimal demo "line-of-business object" type that ezmq can pub/sub. This
type as well as BizFoo showcase just how easily one would extend one's in-house
/ company-specific ezMQ wrapper library by additional custom message types as
desired. In fact, both are ripe candidates for codegen-ing the well-typed
wrapper methods associated with one's line-of-business struct:
`Exchange.PublishXyz`, `Queue.PublishXyz`, and `Queue.SubscribeToXyzs`.

#### func  NewBizEvent

```go
func NewBizEvent(id string, name string) *BizEvent
```

#### func  NewBizEventAt

```go
func NewBizEventAt(id string, name string, date time.Time) *BizEvent
```

#### type BizFoo

```go
type BizFoo struct {
	Bar bool
	Baz int
}
```

Just a minimal demo "line-of-business object" type that ezmq can pub/sub. This
type as well as BizEvent showcase just how easily one would extend one's
in-house / company-specific ezMQ wrapper library by additional custom message
types as desired. In fact, both are ripe candidates for codegen-ing the
well-typed wrapper methods associated with one's line-of-business struct:
`Exchange.PublishXyz`, `Queue.PublishXyz`, and `Queue.SubscribeToXyzs`.

#### func  NewBizFoo

```go
func NewBizFoo(bar bool, baz int) *BizFoo
```

#### type Context

```go
type Context struct {
	UserName string
	Password string
	Host     string
	Port     uint16
}
```

Provides access to the backing message-queue implementation, encapsulating the
underlying connection/channel primitives. Set the fields (as fits the project
context / local setup) before declaring the `Queue`s or `Exchange`s needed to
publish and subscribe, as those calls will connect if the `Context` isn't
already connected. Subsequent field mutations are of course ignored as the
connection is kept alive. For clean-up or manual / pooled connection strategies,
`Context` provides the `Close` method.

#### func (*Context) Close

```go
func (ctx *Context) Close() (chanCloseErr, connCloseErr error)
```
Be SURE to call this when done with ezmq, to cleanly dispose of resources.

#### func (*Context) Exchange

```go
func (ctx *Context) Exchange(name string, cfg *ExchangeConfig, bindTo *Queue) (ex *Exchange, err error)
```
Declares an "exchange" for publishing to multiple subscribers via the specified
`Queue` that MUST have been created with an empty `name`. (NB. if
multiple-subscribers need not be supported, then no need for an `Exchange`: just
use a simple `Queue` only.) If `cfg` is `nil`, the current
`ConfigDefaultsExchange` is used. For `name`, see `Exchange.Name`.

#### func (*Context) Queue

```go
func (ctx *Context) Queue(name string, cfg *QueueConfig) (q *Queue, err error)
```
Declares a queue with the specified `name` for publishing and subscribing. If
`cfg` is `nil`, the current `ConfigDefaultsQueue` is used. For `name`, DO refer
to the docs on `Queue.Name`.

#### type Exchange

```go
type Exchange struct {
	//	Shouldn't be empty, although it's legal: that implies the backing
	//	message-queue's "default" exchange --- however there is typically no
	//	need to set up an `Exchange` to use that one, as it's used by default
	//	for example when `Publish`ing directly via a `Queue`.
	Name string

	//	Set to sensible defaults of `ConfigDefaultsExchange` at initialization.
	Config *ExchangeConfig
}
```

Used in-place-of / in-conjunction-with a `Queue` when needing to publish to
multiple subscribers. ONLY to be constructed via `Context.Exchange()`, and
fields are not to be mutated thereafter! It remains associated with that
`Context` for all its `Publish` calls.

#### func (*Exchange) Publish

```go
func (ex *Exchange) Publish(obj interface{}) error
```
Serializes the specified `obj` to JSON and publishes it to this exchange.

#### func (*Exchange) PublishBizEvent

```go
func (ex *Exchange) PublishBizEvent(evt *BizEvent) error
```
Convenience short-hand for `ex.Publish(evt)`

#### func (*Exchange) PublishBizFoo

```go
func (ex *Exchange) PublishBizFoo(foo *BizFoo) error
```
Convenience short-hand for `ex.Publish(foo)`

#### type ExchangeConfig

```go
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
```

Specialist tweaks for declaring an `Exchange` to the backing message-queue.

#### type Queue

```go
type Queue struct {
	//	Contains either the `name` given at initialization to `Context.Queue()`,
	//	or a random identifier if that was empty: in such cases, this `Queue`
	//	MUST be used to bind to an `Exchange` constructed via `Context.Exchange()`,
	//	and the `Config` will be set to a newly created copy of the effective
	//	`cfg` in `Context.Queue()`, with tweaked `Durable` and `Exclusive` fields.
	Name string

	//	Set to sensible defaults of `ConfigDefaultsQueue` at initialization.
	Config *QueueConfig
}
```

Used to publish and to subscribe. ONLY to be constructed via `Context.Queue()`,
and fields are not to be mutated thereafter! It remains associated with that
`Context` for all its `Publish` and `SubscribeTo` calls.

#### func (*Queue) Publish

```go
func (q *Queue) Publish(obj interface{}) error
```
Serializes the specified `obj` to JSON and publishes it to this exchange.

#### func (*Queue) PublishBizEvent

```go
func (q *Queue) PublishBizEvent(evt *BizEvent) error
```
Convenience short-hand for `q.Publish(evt)`

#### func (*Queue) PublishBizFoo

```go
func (q *Queue) PublishBizFoo(foo *BizFoo) error
```
Convenience short-hand for `q.Publish(foo)`

#### func (*Queue) SubscribeTo

```go
func (q *Queue) SubscribeTo(makeEmptyObjForDeserialization func() interface{}, onMsg func(interface{})) (err error)
```
Generic subscription mechanism used by the more convenient well-typed wrapper
functions such as `SubscribeToEvents` and `SubscribeToFoos`:

Subscribe to messages only of the Type returned by the specified
`makeEmptyObjForDeserialization` constructor function used to allocate a new
empty/zeroed non-pointer struct value whenever attempting to deserialize a
message received from this `Queue`. If that succeeds, a pointer to that value is
passed to the specified `onMsg` event handler: this will always be passed a
non-nil pointer to the value (now populated with data) returned by
`makeEmptyObjForDeserialization`, therefore safe to cast back to the Type.

#### func (*Queue) SubscribeToBizEvents

```go
func (q *Queue) SubscribeToBizEvents(subscribers ...func(*BizEvent)) (err error)
```
A well-typed (to `BizEvent`) wrapper around `Queue.SubscribeTo`.

#### func (*Queue) SubscribeToBizFoos

```go
func (q *Queue) SubscribeToBizFoos(subscribers ...func(*BizFoo)) (err error)
```
A well-typed (to `BizFoo`) wrapper around `Queue.SubscribeTo`.

#### type QueueConfig

```go
type QueueConfig struct {
	Durable                    bool
	AutoDelete                 bool
	Exclusive                  bool
	NoWait                     bool
	Args                       map[string]interface{}
	Pub                        TweakPub
	Sub                        TweakSub
	QosMultipleWorkerInstances bool
}
```

Specialist tweaks for declaring a `Queue` to the backing message-queue.

#### type TweakPub

```go
type TweakPub struct {
	Mandatory  bool
	Immediate  bool
	Persistent bool
}
```

Specialist tweaks for `Publish`ing via a `Queue` or an `Exchange`.

#### type TweakSub

```go
type TweakSub struct {
	Consumer string
	AutoAck  bool
	NoLocal  bool
}
```

Specialist tweaks used from within `Queue.SubscribeTo`.
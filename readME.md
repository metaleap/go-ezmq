# ezmq

```go
    import "github.com/metaleap/go-ezmq"
```

Provides a higher-level, type-driven message-queuing API wrapping RabbitMQ /
amqp.

## High-Level API Workflow:

* make a `Context` (later, when done, `Close` it)
* for **simple messaging**, use it to declare a named `Queue`, then:
    * `Queue.Publish(anything)`
    * `Queue.SubscribeTo<Thing>s(myOnThingHandlers...)`
* to publish to **multiple subscribers**
    * use the `Context` and an unnamed `Queue` to declare an `Exchange`,
    * then `Exchange.Publish(anything)`
* for multiple **worker instances**, set 2 `bool`s, as below

## Example Scenarios

"Line-of-business object" types used here, `BizEvent` and `BizFoo`, are included
for *demo* purposes, to showcase how easily one may "type-safe-ish"ly broadcast
and subscribe-to any kind of custom, in-house struct type.

Pseudo-code ignores all the `error`s returned that it should in real life check:

### Simple publishing via Queue:

```go
    ctx := ezmq.LocalCtx                                        // guest:guest@localhost:5672
    defer ctx.Close()
    var qcfg *ezmq.QueueConfig = nil                            // nil = use 'prudent' defaults
    //
    qe,_ := ctx.Queue('myevents', qcfg)
    qe.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))
    qf,_ := ctx.Queue('myfoos', qcfg)
    qf.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
    // some more for good measure:
    qe.Publish(ezmq.NewBizEvent("evt2", "DatEvent"))
    qf.Publish(&ezmq.BizFoo{ Baz: 20 })                         // same thing just untyped
    qe.Publish(ezmq.NewBizEvent("evt3", "SomeEvent"))           // ditto
```

### Simple subscribing via Queue:

```go
    onBizEvent := func(evt *ezmq.BizEvent) {
        println(evt.Name)
    }
    qe.SubscribeToBizEvents(onBizEvent)
    qf.SubscribeToBizFoos(func(foo *ezmq.Foo) { mylogger.LogAnything(foo) })
    for true { /* we loop until we won't */ }
```

### Multiple subscribers via Exchange:

```go
    qm,_ := ctx.Queue('', qcfg)   //  name MUST be empty
    var xcfg *ezmq.ExchangeConfig = nil // as above, nil = defaults
    ex,_ := ctx.Exchange('mybroadcast', xcfg, qm)  //  only pass `Queue`s that were declared with empty `name`
    ex.Publish(ezmq.NewBizEvent("evt1", "DisEvent"))  //  publish via `Exchange`, not via `Queue`, same API
    ex.Publish(&ezmq.BizFoo{ Bar: true, Baz: 10 })
    ex.Publish(ezmq.NewBizEvent("evt2", "DatEvent")) // same thing just untyped
    ex.Publish(&ezmq.BizFoo{ Baz: 20 }) // ditto
```

### Enabling multiple worker instances:

```go
    // Pass this then to ctx.Queue()
    var qcfg *ezmq.QueueConfig = ezmq.ConfigDefaultsQueue
    qcfg.Pub.Persistent = true
    qcfg.QosMultipleWorkerInstances = true

    // Pass this then to ctx.Exchange(), if one is used (WITH a queue declared with the above)
    var xcfg *ezmq.ExchangeConfig = ezmq.ConfigDefaultsExchange
    xcfg.Pub.Persistent = true

    // Rest as usual
```

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
desired.

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
types as desired.

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
Serializes the specified `obj` and publishes it to this exchange.

#### type ExchangeConfig

```go
type ExchangeConfig struct {
	Type       string // fanout/direct/topic/headers
	Args       map[string]interface{}
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Pub        TweakPub
	QueueBind  struct {
		NoWait     bool
		Args       map[string]interface{}
		RoutingKey string
	}
}
```

Specialist tweaks for declaring an `Exchange` to the backing message-queue. If
you don't know their meaning, you're better off taking our defaults.

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
Serializes the specified `obj` and publishes it to this exchange.

#### func (*Queue) SubscribeTo

```go
func (q *Queue) SubscribeTo(makeEmptyObjForDeserialization func() interface{}, onMsg func(interface{})) (err error)
```
Generic subscription mechanism used by the more convenient well-typed wrapper
functions such as `SubscribeToBizEvents` and `SubscribeToBizFoos`:

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
	QosMultipleWorkerInstances bool
	Pub                        TweakPub
	Sub                        TweakSub
	Args                       map[string]interface{}
}
```

Specialist tweaks for declaring a `Queue` to the backing message-queue. If you
don't know their meaning, you're better off taking our defaults.

#### type TweakPub

```go
type TweakPub struct {
	Mandatory  bool
	Immediate  bool
	Persistent bool
}
```

Specialist tweaks for `Publish`ing via a `Queue` or an `Exchange`. If you don't
know their meaning, you're better off taking our defaults.

#### type TweakSub

```go
type TweakSub struct {
	Consumer string
	AutoAck  bool
	NoLocal  bool

	//	Keep 'nil` to ignore (or set, to handle) unlikely-but-not-impossible
	//	manual-(non-auto)-delivery-acknowledgement errors. RETURN: `true` to
	//	"keep going" (keep listening and also pass the decoded value if any to
	//	subscribers), or `false` to discard the value and stop listening on
	//	behalf of the affected subscribers
	OnAckError func(error) bool
}
```

Specialist tweaks used from within `Queue.SubscribeTo`. If you don't know their
meaning, you're better off taking our defaults.

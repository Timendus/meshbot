package meshwrapper

type EventBody interface {
	Message | Node | ConnectedNode
}

type Event int

const (
	ConnectedEvent Event = iota
	DisconnectedEvent

	AnyMessageEvent
	TextMessageEvent
	NodeInfoEvent
	PositionEvent
	TelemetryEvent
	NeighborInfoEvent
	RoutingEvent
	TraceRouteEvent
	DeviceTelemetryEvent
	EnvironmentTelemetryEvent
	HealthTelemetryEvent
	AirQualityTelemetryEvent
	PowerTelemetryEvent
	LocalStatsTelemetryEvent
)

type pubSub[T EventBody] struct {
	subscriptions map[Event][]func(T)
}

func (ps *pubSub[T]) Subscribe(topic Event, function func(T)) {
	ps.subscriptions[topic] = append(ps.subscriptions[topic], function)
}

func (ps *pubSub[T]) publish(topic Event, msg T) {
	for _, function := range ps.subscriptions[topic] {
		go function(msg)
	}
}

var ConnectionEvents = pubSub[ConnectedNode]{make(map[Event][]func(ConnectedNode))}
var MessageEvents = pubSub[Message]{make(map[Event][]func(Message))}
var NodeEvents = pubSub[Node]{make(map[Event][]func(Node))}

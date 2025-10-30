package meshwrapper

type Event int

const (
	// Connection events
	ConnectedEvent Event = iota
	DisconnectedEvent

	// Message events
	IncomingMessageEvent
	OutgoingMessageEvent

	// Specific message events
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

type EventBody interface {
	IncomingMessage | OutgoingMessage | Node | ConnectedNode
}

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
var IncomingMessageEvents = pubSub[IncomingMessage]{make(map[Event][]func(IncomingMessage))}
var OutgoingMessageEvents = pubSub[OutgoingMessage]{make(map[Event][]func(OutgoingMessage))}
var NodeEvents = pubSub[Node]{make(map[Event][]func(Node))}

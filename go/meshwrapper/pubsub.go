package meshwrapper

type EventBody interface {
	Message | Node | ConnectedNode
}

type pubSub[T EventBody] struct {
	subscriptions map[string][]func(T)
}

func (ps *pubSub[T]) Subscribe(topic string, function func(T)) {
	ps.subscriptions[topic] = append(ps.subscriptions[topic], function)
}

func (ps *pubSub[T]) publish(topic string, msg T) {
	for _, function := range ps.subscriptions[topic] {
		go function(msg)
	}
}

var NodeEvents = pubSub[ConnectedNode]{make(map[string][]func(ConnectedNode))}
var MessageEvents = pubSub[Message]{make(map[string][]func(Message))}

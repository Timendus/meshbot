package meshwrapper

import (
	"log"
	"math/rand/v2"
	"sync/atomic"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

const VERBOSE = false

type acknowledgement struct {
	id        int32
	waiting   atomic.Bool
	delivered chan bool
	repeated  chan bool
	recipient *Node
	status    string
}

// Create a new acknowledgement for a message we're sending to the given node
func newAcknowledgement(node *Node) *acknowledgement {
	ack := acknowledgement{
		id:        rand.Int32(),
		delivered: make(chan bool, 1),
		repeated:  make(chan bool, 1),
		recipient: node,
		waiting:   atomic.Bool{},
		status:    "Waiting",
	}
	ack.waiting.Store(true)
	ack.spam()
	return &ack
}

func (a *acknowledgement) receive(node *Node, err meshtastic.Routing_Error) {
	if !a.waiting.Load() {
		if VERBOSE {
			log.Printf("Acknowledgement %d to %s: Received packed, but was no longer waiting\n", a.id, a.recipient.ColorString())
		}
		return
	}
	if err != meshtastic.Routing_NONE {
		a.negative("Routing error: " + meshtastic.Routing_Error_name[int32(err)])
		return
	}
	if node.Id == a.recipient.Id {
		a.status = "Delivered"
		a.delivered <- true
		a.repeated <- false
		a.close()
	} else {
		a.status = "Repeated"
		a.repeated <- true
		a.spam()
	}
}

func (a *acknowledgement) timeout() {
	a.negative("Timed out")
}

func (a *acknowledgement) error(err error) {
	a.negative("Could not send message: " + err.Error())
}

func (a *acknowledgement) negative(reason string) {
	if !a.waiting.Load() {
		return
	}
	a.status = reason
	a.delivered <- false
	a.repeated <- false
	a.close()
}

func (a *acknowledgement) close() {
	a.waiting.Store(false)
	close(a.delivered)
	close(a.repeated)
	a.spam()
}

func (a *acknowledgement) spam() {
	if VERBOSE {
		log.Printf("Acknowledgement %d to %s: %s\n", a.id, a.recipient.ColorString(), a.status)
	}
}

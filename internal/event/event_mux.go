package event

import (
	"github.com/labstack/gommon/log"
	"github.com/twuillemin/modes-viewer/internal/plane"
)

type PlaneUpdateMultiplexer struct {
	source  chan plane.Plane
	clients map[chan plane.Plane]chan struct{}
}

func CreatePlaneUpdateMultiplexer() PlaneUpdateMultiplexer {

	mux := PlaneUpdateMultiplexer{
		source:  make(chan plane.Plane),
		clients: make(map[chan plane.Plane]chan struct{}),
	}

	go startReception(mux)

	return mux
}

func (mux *PlaneUpdateMultiplexer) GetPublishingQueue() chan plane.Plane {
	return mux.source
}

func (mux *PlaneUpdateMultiplexer) CreateClient() (chan plane.Plane, chan struct{}) {
	client := make(chan plane.Plane)
	stop := make(chan struct{})
	mux.clients[client] = stop
	return client, stop
}

func (mux *PlaneUpdateMultiplexer) RemoveClient(clientChan chan plane.Plane) {

	if clientStop, ok := mux.clients[clientChan]; ok {

		delete(mux.clients, clientChan)

		clientStop <- struct{}{}

		// Close the channels
		close(clientChan)
		close(clientStop)
	}
}

func (mux *PlaneUpdateMultiplexer) Stop() {

	// No more message received
	close(mux.source)

	// Close the client
	for client, _ := range mux.clients {
		mux.RemoveClient(client)
	}
}

func startReception(mux PlaneUpdateMultiplexer) {
	for {
		select {
		case planeEvent := <-mux.source:
			for clientChan, _ := range mux.clients {
				select {
				case clientChan <- planeEvent:
				default:
					log.Print("Unable to send to  client as its queue is full\n")
				}
			}
		}
	}
}

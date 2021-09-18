package tasks

import (
	"log"

	"github.com/reubenmiller/go-c8y/pkg/c8y"

	"github.com/reubenmiller/go-c8y/pkg/microservice"
)

// ExampleHeartbeatTask returns a function which when called creates a event on the given microservice agent managed object
func ExampleHeartbeatTask(ms *microservice.Microservice) func() {
	return func() {
		_, _, err := ms.Client.Event.Create(
			ms.WithServiceUser(),
			c8y.NewEventBuilder(ms.AgentID, "ms_Heartbeat", "Heart beat event"),
		)
		if err != nil {
			log.Printf("Could not create event. %s", err)
		}
	}
}

package gofret

import (
	"testing"
	"time"
)

func TestSendAndReceive(t *testing.T) {
	message := "hello gofret!"
	ready := make(chan bool)
	finished := make(chan bool, 2)

	sender := func() {
		comm := new_communication("localhost:8881")
		if err := comm.Send("localhost:8880", []byte(message)); err != nil {
			t.Errorf("%v", err)
		}
		finished <- true
	}

	receiver := func() {
		comm := new_communication("localhost:8880")
		incoming_messages, err := comm.Listen()
		if err != nil {
			t.Error(err)
		}

		ready <- true

		timeout := time.After(5000 * time.Millisecond)
		select {
		case received_message := <-incoming_messages:
			if string(received_message) != message {
				t.Errorf("expected %v, got %v", message, received_message)
			}
		case <-timeout:
			t.Errorf("didn't receive any message")
		}
		finished <- true
	}

	go receiver()
	<-ready
	go sender()

	<-finished
	<-finished
}

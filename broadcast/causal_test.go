package broadcast

import (
	"testing"
	"time"
)

func TestCausalBroadcast(t *testing.T) {
	peer_addrs := []string{"localhost:9981", "localhost:9982"}
	messages := [][]byte{randStringBytes(2 << 23), []byte("hello gofret!")}
	incoming_channels := []chan []byte{make(chan []byte), make(chan []byte)}
	done_signal := make(chan bool)
	ready_signal := make(chan bool)

	emitter_routine := func(address string, messages [][]byte, incoming_channel chan []byte, done chan bool) {
		broadcaster := CausalBroadcast(Configuration{
			SelfAddress:   address,
			PeerAddresses: peer_addrs,
		})

		incoming_messages, err := broadcaster.Init()
		if err != nil {
			t.Errorf("%v", err)
		}

		go func() {
			for {
				incoming_channel <- <-incoming_messages
			}
		}()

		for _, message := range messages {
			if err := broadcaster.Broadcast(message); err != nil {
				t.Errorf("cannot broadcast: %v", err)
			}
		}

		done <- true
	}

	receiver_routine := func(address string, incoming_channel chan []byte, ready chan bool) {
		broadcaster := CausalBroadcast(Configuration{
			SelfAddress:   address,
			PeerAddresses: peer_addrs,
		})

		incoming_messages, err := broadcaster.Init()
		if err != nil {
			t.Errorf("%v", err)
		}

		go func() {
			for {
				incoming_channel <- <-incoming_messages
			}
		}()
		ready_signal <- true
	}

	go receiver_routine(peer_addrs[0], incoming_channels[0], ready_signal)
	<-ready_signal

	go emitter_routine(peer_addrs[1], messages, incoming_channels[1], done_signal)
	<-done_signal

	timeout := time.After(30000 * time.Millisecond)
	select {
	case incoming_message := <-incoming_channels[0]:
		if string(incoming_message[:len(messages[1])-1]) == string(messages[1]) {
			t.Errorf("order is not correct")
			return
		}
	case incoming_message := <-incoming_channels[1]:
		if string(incoming_message[:len(messages[1])-1]) == string(messages[1]) {
			t.Errorf("order is not correct")
			return
		}
	case <-timeout:
		t.Errorf("timed out")
	}
}

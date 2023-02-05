package broadcast

import (
	"testing"
	"time"
)

func TestUnorderedBroadcast(t *testing.T) {
	peer_addrs := []string{"localhost:8881", "localhost:8882", "localhost:8883", "localhost:8884", "localhost:8885"}
	message := "hello gofret!"

	emitter_routine := func(address string, done chan bool) {
		broadcaster := UnorderedBroadcast(Configuration{
			SelfAddress:   address,
			PeerAddresses: peer_addrs,
		})

		incoming_messages, err := broadcaster.Init()
		if err != nil {
			t.Errorf("%v", err)
		}

		go func() {
			received_message := <-incoming_messages
			if string(received_message) != message {
				t.Errorf("expected %v, got %v", message, received_message)
			}
		}()

		if err := broadcaster.Broadcast([]byte(message)); err != nil {
			t.Errorf("cannot broadcast: %v", err)
		}

		done <- true
	}

	peer_routine := func(address string, done chan bool) {
		broadcaster := UnorderedBroadcast(Configuration{
			SelfAddress:   address,
			PeerAddresses: peer_addrs,
		})

		incoming_messages, err := broadcaster.Init()
		if err != nil {
			t.Errorf("%v", err)
		}

		received_message := <-incoming_messages
		if string(received_message) != message {
			t.Errorf("expected %v, got %v", message, received_message)
		}
		done <- true
	}

	done := make(chan bool, len(peer_addrs)+1)
	for _, peer_addr := range peer_addrs[1:] {
		go peer_routine(peer_addr, done)
	}

	go emitter_routine(peer_addrs[0], done)
	timeout := time.After(5000 * time.Millisecond)

	count := 0
	for {
		select {
		case <-done:
			count += 1
			if count == len(peer_addrs) {
				return
			}
		case <-timeout:
			t.Fatalf("timed out")
		}
	}
}

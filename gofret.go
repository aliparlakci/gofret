package gofret

import (
	"github.com/aliparlakci/gofret/communication"
)

type broadcast_container struct {
	lamport_clock int32
	communication.Communicator
	address          string
	peer_addrs       []string
	delivery_channel chan []byte
}

type Configuration struct {
	SelfAddress   string
	PeerAddresses []string
}

type Broadcaster interface {
	Init() (chan []byte, error)
	Broadcast([]byte) error
}

func (bc *broadcast_container) Connect(address string) {}

func (bc *broadcast_container) Broadcast(message []byte) error {
	bc.delivery_channel <- message
	for _, peer_addr := range bc.peer_addrs {
		if peer_addr == bc.address {
			continue
		}

		if err := bc.Communicator.Send(peer_addr, message); err != nil {
			return err
		}
	}

	return nil
}

func (bc *broadcast_container) Init() (chan []byte, error) {
	bc.Communicator = communication.NewCommunication(bc.address)

	incoming_messages, err := bc.Communicator.Listen()
	if err != nil {
		return nil, err
	}

	bc.delivery_channel = make(chan []byte)
	go bc.handle_new_messages(incoming_messages)
	return bc.delivery_channel, nil
}

func (bc *broadcast_container) handle_new_messages(incoming chan []byte) {
	for {
		message := <-incoming
		bc.delivery_channel <- message
	}
}

func FIFOTotalOrderBroadcast(config Configuration) Broadcaster {
	new_broadcast := broadcast_container{lamport_clock: 0, peer_addrs: config.PeerAddresses, address: config.SelfAddress}
	return &new_broadcast
}

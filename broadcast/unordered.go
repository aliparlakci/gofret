package broadcast

type broadcast_container struct {
	Communicator communicator
	address      string
	peer_addrs   []string
	messages     chan []byte
}

func (bc *broadcast_container) Connect(address string) {}

func (bc *broadcast_container) Broadcast(message []byte) error {
	bc.messages <- message
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
	bc.Communicator = new_communication(bc.address)

	incoming_messages, err := bc.Communicator.Listen()
	if err != nil {
		return nil, err
	}

	bc.messages = incoming_messages
	return incoming_messages, nil
}

func (bc *broadcast_container) Close() error {
	return bc.Communicator.CloseConnection()
}

func UnorderedBroadcast(config Configuration) Broadcaster {
	new_broadcast := broadcast_container{peer_addrs: config.PeerAddresses, address: config.SelfAddress}
	return &new_broadcast
}

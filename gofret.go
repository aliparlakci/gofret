package gofret

import (
	"encoding/json"
	"fmt"
)

type FIFOBroadcaster interface {
	Broadcaster
	Wait() chan bool
}

type fifo_broadcast_message struct {
	Address string // we should find another identifier since address for a node might not be same across nodes
	SendSeq uint64
	Message []byte
}

type fifo_broadcast_container struct {
	config            Configuration
	broadcaster       Broadcaster
	incoming_messages chan []byte
	delivery_channel  chan []byte
	sendSeq           uint64
	delivered         []uint64
	buffer            []fifo_broadcast_message
}

func (f *fifo_broadcast_container) Init() (chan []byte, error) {
	broadcaster := UnorderedBroadcast(f.config)

	var err error
	f.incoming_messages, err = broadcaster.Init()
	if err != nil {
		return nil, err
	}

	f.broadcaster = broadcaster
	f.delivery_channel = make(chan []byte)
	f.sendSeq = 0
	f.buffer = make([]fifo_broadcast_message, 0)
	f.delivered = make([]uint64, len(f.config.PeerAddresses))
	for i := range f.delivered {
		f.delivered[i] = 0
	}

	go f.handle_incoming_messages()
	return f.delivery_channel, nil
}

func (f *fifo_broadcast_container) Broadcast(content []byte) error {
	message := fifo_broadcast_message{
		Address: f.config.SelfAddress,
		Message: content,
		SendSeq: f.sendSeq,
	}
	f.sendSeq++

	marshalled_message, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := f.broadcaster.Broadcast(marshalled_message); err != nil {
		return err
	}
	return nil
}

func (f *fifo_broadcast_container) handle_incoming_messages() {
	for {
		var message fifo_broadcast_message
		data := <-f.incoming_messages
		err := json.Unmarshal(data, &message)
		if err != nil {
			panic(err)
		}

		i, err := f.index_from_address(message.Address)
		if err != nil {
			panic(err)
		}

		if f.delivered[i] <= message.SendSeq {
			f.buffer = append(f.buffer, message)
		}

		f.deliver_messages()
	}
}

func (f *fifo_broadcast_container) deliver_messages() {
	for _, message := range f.buffer {
		i, err := f.index_from_address(message.Address)
		if err != nil {
			panic(err)
		}

		if f.delivered[i] == message.SendSeq {
			f.delivery_channel <- message.Message
			f.delivered[i]++
		}
	}

	f.cleanup_buffer()
}

func (f *fifo_broadcast_container) cleanup_buffer() {
	new_buffer := make([]fifo_broadcast_message, 0)

	for _, message := range f.buffer {
		i, err := f.index_from_address(message.Address)
		if err != nil {
			panic(err)
		}

		if f.delivered[i] <= message.SendSeq {
			new_buffer = append(new_buffer, message)
		}
	}

	f.buffer = new_buffer
}

func (f *fifo_broadcast_container) index_from_address(searched_address string) (uint, error) {
	for i, address := range f.config.PeerAddresses {
		if address == searched_address {
			return uint(i), nil
		}
	}
	return 0, fmt.Errorf("cannot find address %v", searched_address)
}

func FIFOBroadcast(config Configuration) Broadcaster {
	fifo_broadcaster := fifo_broadcast_container{config: config}
	return &fifo_broadcaster
}

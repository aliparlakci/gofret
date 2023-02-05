package broadcast

import (
	"encoding/json"
)

type dependency_map map[string]uint64

type causal_broadcast_message struct {
	Address      string // we should find another identifier since address for a node might not be same across nodes
	Dependencies dependency_map
	Message      []byte
}

type causal_broadcast_container struct {
	config            Configuration
	broadcaster       Broadcaster
	incoming_messages chan []byte
	delivery_channel  chan []byte
	sendSeq           uint64
	delivered         dependency_map
	buffer            []causal_broadcast_message
}

func (d dependency_map) NotGreaterThan(another_deps dependency_map) bool {
	for key, value := range d {
		if value > another_deps[key] {
			return false
		}
	}
	return true
}

func (c *causal_broadcast_container) Init() (chan []byte, error) {
	broadcaster := UnorderedBroadcast(c.config)

	var err error
	c.incoming_messages, err = broadcaster.Init()
	if err != nil {
		return nil, err
	}

	c.broadcaster = broadcaster
	c.delivery_channel = make(chan []byte)
	c.sendSeq = 0
	c.buffer = make([]causal_broadcast_message, 0)
	c.delivered = make(dependency_map, len(c.config.PeerAddresses))
	for _, address := range c.config.PeerAddresses {
		c.delivered[address] = 0
	}

	go c.handle_incoming_messages()
	return c.delivery_channel, nil
}

func (c *causal_broadcast_container) Broadcast(content []byte) error {
	dependencies := make(map[string]uint64, len(c.config.PeerAddresses))
	for address, seq := range c.delivered {
		dependencies[address] = seq
	}
	dependencies[c.config.SelfAddress] = c.sendSeq

	message := causal_broadcast_message{
		Address:      c.config.SelfAddress,
		Message:      content,
		Dependencies: dependencies,
	}
	c.sendSeq++

	marshalled_message, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := c.broadcaster.Broadcast(marshalled_message); err != nil {
		return err
	}
	return nil
}

func (c *causal_broadcast_container) handle_incoming_messages() {
	for {
		var message causal_broadcast_message
		data := <-c.incoming_messages
		err := json.Unmarshal(data, &message)
		if err != nil {
			panic(err)
		}

		c.buffer = append(c.buffer, message)

		c.deliver_messages()
	}
}
func (c *causal_broadcast_container) deliver_messages() {
	new_buffer := make([]causal_broadcast_message, 0)
	for _, message := range c.buffer {
		if message.Dependencies.NotGreaterThan(c.delivered) {
			c.delivery_channel <- message.Message
			c.delivered[message.Address]++
		}
		new_buffer = append(new_buffer, message)
	}

	c.buffer = new_buffer
}

func CausalBroadcast(config Configuration) Broadcaster {
	causal_broadcaster := causal_broadcast_container{config: config}
	return &causal_broadcaster
}

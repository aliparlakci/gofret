package broadcast

type Configuration struct {
	SelfAddress   string
	PeerAddresses []string
}

type Broadcaster interface {
	Init() (chan []byte, error)
	Broadcast([]byte) error
	Close() error
}

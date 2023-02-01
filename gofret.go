package gofret

type Broadcast struct {
}

func (b *Broadcast) Connect(address string) {}

func (b *Broadcast) Broadcast(input []byte) error {
	return nil
}

func FIFOTotalOrderBroadcast() Broadcast {
	return Broadcast{}
}

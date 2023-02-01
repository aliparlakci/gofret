package gofret_test

import (
	"testing"

	"github.com/aliparlakci/gofret"
)

func TestFIFOTotalOrder(t *testing.T) {
	broadcaster := gofret.FIFOTotalOrderBroadcast()
	broadcaster.Connect("tcp://localhost:8181/")

	msg := []byte("hello")
	if err := broadcaster.Broadcast(msg); err != nil {
		panic(err)
	}
}

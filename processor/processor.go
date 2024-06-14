package processor

import (
	"context"
	"gossip-protocol/network"
	"sync"
)

func Processor(ctx context.Context, wg *sync.WaitGroup, obsReceiveRes <-chan network.Message, tracker network.MessageTracker) []byte {
	defer wg.Done()
	for {
		select {
		case m := <-obsReceiveRes:
			tracker.Add(&m)
		case <-ctx.Done():
			return nil
		}
	}
}
package whatsmeow

import (
	"context"
	"errors"
	waBinary "github.com/polymorfa/hypermeow/binary"
	"github.com/polymorfa/hypermeow/types"
	"sync/atomic"
	"testing"
	"time"
)

func abpropsNotification() *waBinary.Node {
	return &waBinary.Node{Tag: "notification", Attrs: waBinary.Attrs{"type": "server", "id": "1", "from": types.ServerJID}, Content: []waBinary.Node{{Tag: "abprops"}}}
}
func TestABPropsAckWaitsWithoutBlockingReceiveLoop(t *testing.T) {
	started := make(chan struct{})
	complete := make(chan struct{})
	acked := make(chan struct{}, 1)
	var tasks atomic.Int32
	cli := &Client{ServerABPropsSync: func(context.Context) error { tasks.Add(1); close(started); <-complete; return nil }}
	node := abpropsNotification()
	if !cli.deferServerABProps(context.Background(), node, func() { acked <- struct{}{} }) {
		t.Fatal("not deferred")
	}
	<-started
	if !cli.deferServerABProps(context.Background(), node, func() { t.Error("duplicate ack") }) {
		t.Fatal("duplicate not coalesced")
	}
	select {
	case <-acked:
		t.Fatal("acked before task/IQ completion")
	default:
	}
	close(complete)
	select {
	case <-acked:
	case <-time.After(time.Second):
		t.Fatal("missing ack")
	}
	if tasks.Load() != 1 {
		t.Fatal("duplicate task")
	}
}
func TestABPropsAbsentUnrelatedFailedAndCancelled(t *testing.T) {
	if (&Client{}).deferServerABProps(context.Background(), abpropsNotification(), nil) {
		t.Fatal("nil handler changed path")
	}
	for _, failure := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if !failure {
			cancel()
		}
		done := make(chan struct{})
		cli := &Client{ServerABPropsSync: func(context.Context) error { defer close(done); return errors.New("incomplete") }}
		if !cli.deferServerABProps(ctx, abpropsNotification(), func() { t.Error("failed/cancelled task acked") }) {
			t.Fatal("not deferred")
		}
		if failure {
			<-done
		}
		cancel()
	}
	cli := &Client{ServerABPropsSync: func(context.Context) error { t.Fatal("unrelated callback"); return nil }}
	node := abpropsNotification()
	node.Attrs["type"] = "encrypt"
	if cli.deferServerABProps(context.Background(), node, nil) {
		t.Fatal("unrelated deferred")
	}
	node = abpropsNotification()
	node.Content = nil
	if cli.deferServerABProps(context.Background(), node, nil) {
		t.Fatal("missing child deferred")
	}
}

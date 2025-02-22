//go:build simtest

package testlistener

import (
	"errors"
	"net"
	"testing"
	"testing/synctest"
)

func TestListener_closeDuringConnect(t *testing.T) {
	synctest.Run(func() {
		testListener_closeDuringConnect(t)
	})
}

func testListener_closeDuringConnect(t *testing.T) {
	l := NewListener()

	connectErrs := make(chan error, 1)
	go func() {
		_, err := l.Connect()
		connectErrs <- err
	}()

	// Wait for the goroutines in the bubble to block
	synctest.Wait()

	if err := l.Close(); err != nil {
		t.Errorf("Close listener failed: %v", err)
	}

	if err := <-connectErrs; !errors.Is(err, net.ErrClosed) {
		t.Errorf("blocked Connect got error %v, want %v", err, net.ErrClosed)
	}
}

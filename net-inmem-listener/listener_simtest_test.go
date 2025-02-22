//go:build simtest

package testlistener

import (
	"errors"
	"net"
	"testing"
	"testing/synctest"
)

func TestListener_CloseDuringAccept(t *testing.T) {
	synctest.Run(func() {
		l := NewListener()

		errs := make(chan error, 1)
		go func() {
			_, err := l.Accept()
			errs <- err
		}()

		// Wait for the accept goroutine to block the bubble
		synctest.Wait()

		// Close the listener
		if err := l.Close(); err != nil {
			t.Errorf("Close listener failed: %v", err)
		}

		if err := <-errs; !errors.Is(err, net.ErrClosed) {
			t.Errorf("blocked Accept got error %v, want %v", err, net.ErrClosed)
		}
	})
}

func TestListener_CloseDuringConnect(t *testing.T) {
	synctest.Run(func() {
		l := NewListener()

		errs := make(chan error, 1)
		go func() {
			_, err := l.Connect()
			errs <- err
		}()

		// Wait for the connect goroutine to block the bubble
		synctest.Wait()

		if err := l.Close(); err != nil {
			t.Errorf("Close listener failed: %v", err)
		}

		if err := <-errs; !errors.Is(err, net.ErrClosed) {
			t.Errorf("blocked Connect got error %v, want %v", err, net.ErrClosed)
		}
	})
}

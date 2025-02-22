package testlistener

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestListener_AcceptConnect(t *testing.T) {
	l := NewListener()
	t.Cleanup(func() {
		l.Close()
	})

	acceptDone := make(chan struct{})
	var serverConn net.Conn
	var acceptErr error
	go func() {
		serverConn, acceptErr = l.Accept()
		close(acceptDone)
	}()

	clientConn, err := l.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for accept to complete
	<-acceptDone
	if acceptErr != nil {
		t.Fatalf("Accept failed: %v", acceptErr)
	}

	t.Cleanup(func() {
		serverConn.Close()
	})

	// Test data transfer
	testData := []byte("hello")
	go func() {
		clientConn.Write(testData)
		clientConn.Close()
	}()

	buf := make([]byte, len(testData))
	n, err := io.ReadFull(serverConn, buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Read %d bytes, want %d", n, len(testData))
	}
	if string(buf) != string(testData) {
		t.Errorf("Read %q, want %q", buf, testData)
	}
}

func TestListener_Close(t *testing.T) {
	l := NewListener()

	if err := l.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify that Accept returns net.ErrClosed after close
	_, err := l.Accept()
	if !errors.Is(err, net.ErrClosed) {
		t.Errorf("Accept after close got error %v, want %v", err, net.ErrClosed)
	}

	// Verify that Connect returns net.ErrClosed after close
	_, err = l.Connect()
	if !errors.Is(err, net.ErrClosed) {
		t.Errorf("Connect after close got error %v, want %v", err, net.ErrClosed)
	}
}

func TestListener_AcceptTimeout(t *testing.T) {
	l := NewListener()

	// Create a channel to signal the accept goroutine is running
	accepting := make(chan struct{})
	acceptDone := make(chan error, 1)

	// Start an accept operation
	go func() {
		close(accepting) // Signal that we're about to call Accept
		conn, err := l.Accept()
		if conn != nil {
			conn.Close()
		}
		acceptDone <- err
	}()

	// Wait for the accept goroutine to start
	<-accepting

	// Close the listener
	if err := l.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Wait for accept to complete
	select {
	case err := <-acceptDone:
		if err != net.ErrClosed {
			t.Errorf("Accept after close got error %v, want %v", err, net.ErrClosed)
		}
	case <-time.After(time.Second):
		t.Error("Accept did not return after listener close")
	}
}

func TestListener_multipleConcurrentConnections(t *testing.T) {
	l := NewListener()
	t.Cleanup(func() {
		l.Close()
	})

	const numConns = 5
	var wg sync.WaitGroup
	wg.Add(numConns * 2) // For both client and server goroutines

	// Start accepting connections
	for range numConns {
		go func() {
			defer wg.Done()
			conn, err := l.Accept()
			if err != nil {
				t.Errorf("Accept failed: %v", err)
				return
			}
			defer conn.Close()

			// Read the connection index
			buf := make([]byte, 1)
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Errorf("Read failed: %v", err)
			}
		}()
	}

	// Create multiple connections
	for i := range numConns {
		go func(idx int) {
			defer wg.Done()
			conn, err := l.Connect()
			if err != nil {
				t.Errorf("Connect failed: %v", err)
				return
			}
			defer conn.Close()

			// Write the connection index
			if _, err := conn.Write([]byte{byte(idx)}); err != nil {
				t.Errorf("Write failed: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

func TestListener_Addr(t *testing.T) {
	l := NewListener()
	t.Cleanup(func() {
		l.Close()
	})

	addr := l.Addr()
	if got := addr.Network(); got != "addr" {
		t.Errorf("addr.Network() = %q, want \"addr\"", got)
	}
	if got := addr.String(); got != "addr" {
		t.Errorf("addr.String() = %q, want \"addr\"", got)
	}
}

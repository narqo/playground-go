package testlistener

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"
)

func TestListener_AcceptConnect(t *testing.T) {
	l := NewListener()
	t.Cleanup(func() {
		l.Close()
	})

	acceptErrs := make(chan error, 1)
	var serverConn net.Conn
	go func() {
		var err error
		serverConn, err = l.Accept()
		acceptErrs <- err
	}()

	clientConn, err := l.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait until the accept goroutine completes
	if err := <-acceptErrs; err != nil {
		t.Fatalf("Accept failed: %v", err)
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
		t.Fatalf("Read from server connection failed: %v", err)
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

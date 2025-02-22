package testlistener

import "net"

// Listener implements net.Listener.
type Listener struct {
	addr  addr
	conns chan net.Conn
	done  chan struct{}
}

// addr implements net.Addr.
type addr struct{}

func (a addr) Network() string { return "addr" }
func (a addr) String() string  { return "addr" }

// NewListener creates a new in-memory listener.
func NewListener() *Listener {
	return &Listener{
		addr:  addr{},
		conns: make(chan net.Conn),
		done:  make(chan struct{}),
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conns:
		return conn, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *Listener) Close() error {
	close(l.done)
	return nil
}

func (l *Listener) Addr() net.Addr {
	return l.addr
}

// Connect creates a new connection pair and sends one end to the listener.
func (l *Listener) Connect() (net.Conn, error) {
	// Check if listener is closed first
	select {
	case <-l.done:
		return nil, net.ErrClosed
	default:
	}

	client, server := net.Pipe()
	select {
	case l.conns <- server:
		return client, nil
	case <-l.done:
		client.Close()
		server.Close()
		return nil, net.ErrClosed
	}
}

package main

import (
	"time"
	"context"
	"net"

	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)


type DtnStream struct {

}

func (ds *DtnStream) Write(p []byte) (n int, err error) {

	return 0, nil
	}

func (ds *DtnStream) Read(p []byte) (n int, err error) {
	return 0, nil

	}


func (ds *DtnStream) 	Close() error {
	return nil
}


// Reset closes both ends of the stream. Use this to tell the remote
// side to hang up and go away.

func (ds *DtnStream)  Reset() error {
	return nil
}


func (ds *DtnStream) SetDeadline(time.Time) error {
	return nil
}

func (ds *DtnStream) SetReadDeadline(time.Time) error {
	return nil
}

func (ds *DtnStream) SetWriteDeadline(time.Time) error{
	return nil
}




type DtnConn struct {}
func (dc *DtnConn) 	Close() error {
	return nil
}

// IsClosed returns whether a connection is fully closed, so it can
// be garbage collected.
func (dc *DtnConn)  IsClosed() bool {
	return false
}

// OpenStream creates a new stream.
func (dc *DtnConn)  OpenStream() (DtnStream, error) {
	return DtnStream{}, nil
	}

// AcceptStream accepts a stream opened by the other side.
func (dc *DtnConn) AcceptStream()  (DtnStream, error) {
	return DtnStream{}, nil
}


type DtnAddr struct {}

func (da DtnAddr) Network() string {
	return ""
}
func (da DtnAddr) String() string {
	return ""
}

type DtnListener struct {}

func (dn *DtnListener) Accept() (DtnConn, error) {
	return DtnConn{}, nil
}


func (dn *DtnListener) Close() error {
	return nil
}


func (dn *DtnListener) Addr() net.Addr {
	return DtnAddr{}
}


type DtnTransport struct {}

func NewDtnTransport() *DtnTransport {
	return &DtnTransport{}
}


func (dt *DtnTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (DtnConn, error) {
	return DtnConn{}, nil
}

// CanDial returns true if this transport knows how to dial the given
// multiaddr.
//
// Returning true does not guarantee that dialing this multiaddr will
// succeed. This function should *only* be used to preemptively filter
// out addresses that we can't dial.
func (dt *DtnTransport) CanDial(addr ma.Multiaddr) bool {
	return false
}

// Listen listens on the passed multiaddr.
func (dt *DtnTransport) Listen(laddr ma.Multiaddr) (DtnListener, error) {
	return DtnListener{}, nil
}

// Protocol returns the set of protocols handled by this transport.
//
// See the Network interface for an explanation of how this is used.
func (dt *DtnTransport) Protocols() []int {
	return []int{}
}

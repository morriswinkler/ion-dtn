package dtn

import (
	"context"
	"log"
	"net"

	peer "github.com/libp2p/go-libp2p-peer"
	tpt "github.com/libp2p/go-libp2p-transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	ma "github.com/multiformats/go-multiaddr"
)

type DtnTransport struct {
	Upgrader *tptu.Upgrader
}

func New(u *tptu.Upgrader) *DtnTransport {
	return &DtnTransport{u}
}

func (c *DtnConn) Dial(network, address string) (Conn, Error) {
}

func (t *DtnTransport) Dial(ctx context.Context, addr ma.Multiaddr, p peer.ID) (tpt.Conn, error) {
}

func (t *DtnTransport) CanDial(addr ma.Multiaddr) bool {
	// TODO implement
	return true
}

func (t *DtnTransport) Listen(addr ma.Multiaddr) (tpt.Listener, error) {
}

type dialer struct{}

package netadapter

import (
	"fmt"

	"github.com/kaspanet/kaspad/domainmessage"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/netadapter/server"
)

// NetConnection is a wrapper to a server connection for use by services external to NetAdapter
type NetConnection struct {
	connection            server.Connection
	id                    *id.ID
	router                *routerpkg.Router
	onDisconnectedHandler server.OnDisconnectedHandler
}

func newNetConnection(connection server.Connection, routerInitializer RouterInitializer) *NetConnection {
	router := routerpkg.NewRouter()

	netConnection := &NetConnection{
		connection: connection,
		router:     router,
	}

	netConnection.connection.SetOnDisconnectedHandler(func() {
		router.Close()

		if netConnection.onDisconnectedHandler != nil {
			netConnection.onDisconnectedHandler()
		}
	})

	router.SetOnRouteCapacityReachedHandler(func() {
		netConnection.Disconnect()
	})

	routerInitializer(router, netConnection)

	return netConnection
}

func (c *NetConnection) start() {
	c.connection.Start(c.router)
}

func (c *NetConnection) String() string {
	return fmt.Sprintf("<%s: %s>", c.id, c.connection)
}

// ID returns the ID associated with this connection
func (c *NetConnection) ID() *id.ID {
	return c.id
}

// SetID sets the ID associated with this connection
func (c *NetConnection) SetID(peerID *id.ID) {
	c.id = peerID
}

// Address returns the address associated with this connection
func (c *NetConnection) Address() string {
	return c.connection.Address().String()
}

// IsOutbound returns whether the connection is outbound
func (c *NetConnection) IsOutbound() bool {
	return c.connection.IsOutbound()
}

// NetAddress returns the NetAddress associated with this connection
func (c *NetConnection) NetAddress() *domainmessage.NetAddress {
	return domainmessage.NewNetAddress(c.connection.Address(), 0)
}

// SetOnInvalidMessageHandler sets a handler function
// for invalid messages
func (c *NetConnection) SetOnInvalidMessageHandler(onInvalidMessageHandler server.OnInvalidMessageHandler) {
	c.connection.SetOnInvalidMessageHandler(onInvalidMessageHandler)
}

func (c *NetConnection) setOnDisconnectedHandler(onDisconnectedHandler server.OnDisconnectedHandler) {
	c.onDisconnectedHandler = onDisconnectedHandler
}

// Disconnect disconnects the given connection
func (c *NetConnection) Disconnect() {
	c.connection.Disconnect()
}

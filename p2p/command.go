package p2p

import (
	"net"

	"github.com/FactomProject/factomd/common/primitives"
)

type P2PCommand interface{}

// CommandDialPeer is used to instruct the Controller to dial a peer address
type CommandDialPeer struct {
	persistent bool
	peer       Peer
}

// CommandAddPeer is used to instruct the Controller to add a connection
// This connection can come from acceptLoop or some other way.
type CommandAddPeer struct {
	conn net.Conn
}

// CommandShutdown is used to instruct the Controller to takve various actions.
type CommandShutdown struct {
	_ uint8
}

// CommandAdjustPeerQuality is used to instruct the Controller to reduce a connections quality score
type CommandAdjustPeerQuality struct {
	PeerHash   string
	Adjustment int32
}

func (e *CommandAdjustPeerQuality) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandAdjustPeerQuality) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandAdjustPeerQuality) String() string {
	str, _ := e.JSONString()
	return str
}

// CommandBan is used to instruct the Controller to disconnect and ban a peer
type CommandBan struct {
	PeerHash string
}

func (e *CommandBan) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandBan) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandBan) String() string {
	str, _ := e.JSONString()
	return str
}

// CommandDisconnect is used to instruct the Controller to disconnect from a peer
type CommandDisconnect struct {
	PeerHash string
}

func (e *CommandDisconnect) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandDisconnect) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandDisconnect) String() string {
	str, _ := e.JSONString()
	return str
}

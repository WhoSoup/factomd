package networkcontrol

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// The actions corresponding to ExtID[0]
const (
	ActionPromote = iota
	ActionRemove
	ActionVote
)

// ProposalDuration is the amount of time (in blocks) that a proposal lasts before expiring
var ProposalDuration = 1000

// ChainID is the chain all proposals and votes are submitted to
var ChainID = "888888165d185ba3342d8f0dcc331066f454196f1ad7060b00f856b6f483b619"

// ChainIDBytes is the chain all proposals and votes are submitted to
var ChainIDBytes = []byte{0x88, 0x88, 0x88, 0x16, 0x5d, 0x18, 0x5b, 0xa3, 0x34, 0x2d, 0x8f, 0x0d, 0xcc, 0x33, 0x10, 0x66, 0xf4, 0x54, 0x19, 0x6f, 0x1a, 0xd7, 0x06, 0x0b, 0x00, 0xf8, 0x56, 0xb6, 0xf4, 0x83, 0xb6, 0x19}

// ChainIDHash is the chain all proposals and votes are submitted to
var ChainIDHash interfaces.IHash

func init() {
	ChainIDHash = primitives.NewHash(ChainIDBytes)
}

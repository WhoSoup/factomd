package networkcontrol

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Manager struct {
	state interfaces.IState // temporary to figure out what is needed
}

func NewManager(state interfaces.IState) *Manager {
	m := new(Manager)
	m.state = state
	return m
}

func (m *Manager) ParseEntry(entry interfaces.IEBEntry) {
	extids := entry.ExternalIDs()
	action, _ := primitives.DecodeVarInt(extids[0])
	switch action {
	case ActionPromote:
	case ActionRemove:
	case ActionVote:
	default:
		return
	}
}

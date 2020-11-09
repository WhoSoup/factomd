package networkcontrol

import (
	"errors"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Manager struct {
	state interfaces.IState // temporary to figure out what is needed

	// it's a slice because we're not expecting many proposals to exist simltaneously
	proposals []proposal
}

func NewManager(state interfaces.IState) *Manager {
	m := new(Manager)
	m.state = state
	return m
}

func (m *Manager) ParseEntry(dblock interfaces.IDirectoryBlock, entry interfaces.IEBEntry) {
	extids := entry.ExternalIDs()
	action, _ := primitives.DecodeVarInt(extids[0])
	switch action {
	case ActionPromoteAudit:
		m.verifyProposal(dblock.GetTimestamp(), entry)
	case ActionPromoteFed:
	case ActionRemove:
	case ActionVote:
	default:
		return
	}
}

// extid[0] (action) already verified
func (m *Manager) verifyProposal(time interfaces.Timestamp, entry interfaces.IEBEntry) error {
	extids := entry.ExternalIDs()

	var ts primitives.Timestamp
	if err := ts.UnmarshalBinary(extids[1]); err != nil {
		return err
	}

	diff := time.GetTimeSeconds() - ts.GetTimeSeconds()
	if diff < -3600 || diff > 3600 {
		return errors.New("timestamp outside of range")
	}

	var id, target primitives.Hash
	var sig primitives.Signature

	if err := id.SetBytes(extids[2]); err != nil {
		return err
	}

	if err := sig.UnmarshalBinary(append(extids[3], entry.GetContent()...)); err != nil {
		return err
	}

	if err := target.SetBytes(extids[4]); err != nil {
		return err
	}

	// check if part of auth set
	var server interfaces.IServer

	feds := m.state.GetFedServers(m.state.GetLLeaderHeight())
	for _, f := range feds {
		if f.GetChainID().IsSameAs(id.GetHash()) {
			server = f
			break
		}
	}

	if server == nil {
		audits := m.state.GetAuditServers(m.state.GetLLeaderHeight())
		for _, a := range audits {
			if a.GetChainID().IsSameAs(id.GetHash()) {
				server = a
				break
			}
		}
	}

	if server == nil {
		return errors.New("proposal submitted by server not in auth set")
	}

	// todo verify pubkey
	// todo verify sig

	return nil
}

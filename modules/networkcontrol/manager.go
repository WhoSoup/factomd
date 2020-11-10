package networkcontrol

import (
	"bytes"
	"errors"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Manager struct {
	state interfaces.IState // temporary to figure out what is needed

	// it's a slice because we're not expecting many proposals to exist simltaneously
	proposals []*proposal
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
		if err := m.verifyProposal(dblock.GetTimestamp(), entry); err != nil {
			// TODO log warning?
			return
		}
		prop := newProposal(entry, action)
		m.proposals = append(m.proposals, prop)

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

	if len(extids) != 5 {
		return errors.New("invalid amount of extids")
	}

	var ts primitives.Timestamp
	if err := ts.UnmarshalBinary(extids[1]); err != nil {
		return err
	}

	diff := time.GetTimeSeconds() - ts.GetTimeSeconds()
	if diff < -3600 || diff > 3600 {
		return errors.New("timestamp outside of range")
	}

	var id, target primitives.Hash
	if err := id.SetBytes(extids[2]); err != nil {
		return err
	}

	sig := new(primitives.Signature)
	if err := sig.UnmarshalBinary(append(extids[3], entry.GetContent()...)); err != nil {
		return err
	}

	if err := target.SetBytes(extids[4]); err != nil {
		return err
	}

	var message bytes.Buffer
	for _, e := range extids {
		if _, err := message.Write(e); err != nil {
			return err
		}
	}

	status, err := m.state.FastVerifyAuthoritySignature(message.Bytes(), sig, m.state.GetLLeaderHeight())
	if err != nil { // invalid sig
		return err
	}

	// audit or fed ok
	if status < 0 {
		return errors.New("not signed by an authority")
	}

	return nil
}

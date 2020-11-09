package networkcontrol

import (
	"sync/atomic"

	"github.com/FactomProject/factomd/common/interfaces"
)

type proposal struct {
	action    uint64
	entryHash interfaces.IHash
	count     uint32
}

func newProposal(ehash interfaces.IHash, action uint64) *proposal {
	p := new(proposal)
	p.entryHash = ehash
	p.action = action
	return p
}

func (p *proposal) Add() {
	atomic.AddUint32(&p.count, 1)
}

func (p *proposal) Count() uint32 {
	return p.count
}

func (p *proposal) BuildMessage() interfaces.IMsg {
	return nil
}

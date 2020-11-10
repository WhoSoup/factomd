package networkcontrol

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
)

type proposal struct {
	action    uint64
	entryHash interfaces.IHash

	voteMtx sync.RWMutex
	vote    map[string]bool
}

func newProposal(entry interfaces.IEBEntry, action uint64) *proposal {
	p := new(proposal)
	p.entryHash = entry.GetHash()
	p.action = action
	p.vote = make(map[string]bool)
	p.vote[fmt.Sprintf("%x", entry.ExternalIDs()[2])] = true
	return p
}

func (p *proposal) Add(chain string) {
	p.voteMtx.Lock()
	p.vote[chain] = true
	p.voteMtx.Unlock()
}

func (p *proposal) Count() int {
	return len(p.vote)
}

func (p *proposal) BuildMessage() interfaces.IMsg {
	return nil
}

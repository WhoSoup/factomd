package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

const window = 2000
const askLimit = 10
const timeout = time.Second * 30

type BetterCatchup struct {
	floor   uint32
	ceiling uint32

	asked   map[uint32]time.Time
	arrived map[uint32]bool

	mtx sync.Mutex
}

func init() {

}

type CatchupStatus uint8

const (
	CatchupArrived = iota
	CatchupSynced
	CatchupFailed
)

type CatchupNotice struct {
	Height uint32
	Status CatchupStatus
}

func (dbsl *DBStateList) BetterCatchup() {
	s := dbsl.State

	// get the height of the known blocks
	ceiling := func() (rval uint32) {
		a := dbsl.State.GetHighestAck()
		k := dbsl.State.GetHighestKnownBlock()
		// check that known is more than 2 ahead of acknowledged to make
		// sure not to ask for blocks that haven't finished
		if k > a+2 {
			return k - 2
		}
		if a < 2 {
			return a
		}
		return a - 2 // Acks are for height + 1 (sometimes +2 in min 0)
	}

	bc := new(BetterCatchup)
	bc.arrived = make(map[uint32]bool)
	bc.asked = make(map[uint32]time.Time)
	bc.floor = s.highestKnown
	bc.raiseCeiling(ceiling())

	fmt.Printf("[BCU] starting from %d to %d\n", bc.floor, bc.ceiling)

	// mark arrivals
	go bc.HandleNotices(s.CatchupNotify)

	// Don't start until the db is finished loading.
	for !s.DBFinished {
		time.Sleep(time.Second)
	}
	if s.highestKnown < s.DBHeightAtBoot {
		s.highestKnown = s.DBHeightAtBoot + 1 // Make sure we ask for the next block after the database at startup.
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			bc.raiseCeiling(ceiling())
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			bc.send(s)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for range ticker.C {
			bc.mtx.Lock()
			fmt.Printf("%+v\n", bc)
			bc.mtx.Unlock()
		}
	}()
}

func (bc *BetterCatchup) HandleNotices(in <-chan CatchupNotice) {
	for n := range in {
		fmt.Println("[BCU] notice:", n.Height, n.Status)
		switch n.Status {
		case CatchupArrived:
			bc.MarkAsArrived(n.Height)
		case CatchupFailed:
			bc.MarkAsFailed(n.Height)
		case CatchupSynced:
			bc.MarkAsSynced(n.Height)
		}
	}
}

func (bc *BetterCatchup) MarkAsArrived(height uint32) {
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	bc.arrived[height] = true
	delete(bc.asked, height)

	fmt.Println("[BCU] Marked as Arrived:", height)
}

func (bc *BetterCatchup) MarkAsSynced(height uint32) {
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	if height <= bc.floor {
		return
	}
	for i := bc.floor; i <= height; i++ {
		delete(bc.asked, i)
		delete(bc.arrived, i)
	}
	bc.floor = height + 1
	fmt.Println("[BCU] Marked as Synced:", height)
}

func (bc *BetterCatchup) raiseCeiling(height uint32) {
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	if bc.ceiling > height {
		return
	}
	if bc.floor+window > height {
		bc.ceiling = height
	} else {
		bc.ceiling = bc.floor + window
	}
	fmt.Println("[BCU] Ceiling raised to:", bc.ceiling)
}

func (bc *BetterCatchup) MarkAsFailed(height uint32) {
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	if height <= bc.floor {
		return
	}
	bc.asked[height] = time.Time{}
	delete(bc.arrived, height)
	fmt.Println("[BCU] Marked as Failed:", height)
}

func (bc *BetterCatchup) send(state interfaces.IState) {
	bc.mtx.Lock()
	defer bc.mtx.Unlock()

	sendable := make([]uint32, 0)
	for i := bc.floor; i <= bc.ceiling; i++ {
		if bc.canSend(i) {
			sendable = append(sendable, i)
		}
	}

	slots := partition(sendable)
	for _, slot := range slots {
		msg := messages.NewDBStateMissing(state, slot[0], slot[1])
		msg.SendOut(state, msg)
		fmt.Println("[BCU] Sending for:", slot[0], "to", slot[1])
	}
}

// takes ORDERED input of a,a+1,a+2,a+3,..,a+n,b,b+1,...b+m,c...c+o
// where b > a+n+1 and c > b+m+1
// a slice of (start, end) tuples, ie [[a,a+n],[b,b+m],[c,c+o]]
func partition(n []uint32) [][]uint32 {
	res := make([][]uint32, 0)
	var seq []uint32
	for i, v := range n {
		if len(seq) == 0 {
			seq = append(seq, v)
		}
		if len(seq) > 0 && (i == len(n)-1 || n[i+1] != v+1 || v == seq[0]+askLimit) {
			seq = append(seq, v)
			res = append(res, seq)
			seq = nil
		}
	}

	return res
}

// only use inside a mutex
func (bc *BetterCatchup) canSend(i uint32) bool {
	if i >= bc.ceiling || i < bc.floor {
		return false
	}
	if t, ok := bc.asked[i]; !ok || (time.Since(t) >= timeout && !bc.arrived[i]) {
		return true
	}
	return false
}

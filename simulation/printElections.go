package simulation

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/fnode"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
)

func lookup(id interfaces.IHash) *state.State {
	for _, fn := range fnode.GetFnodes() {
		if fn.State.IdentityChainID.Fixed() == id.Fixed() {
			return fn.State
		}
	}
	return nil
}

func printSimElections(elects *int, value int, listenTo *int, wsapiNode *int) {
	out := ""

	if *listenTo < 0 || *listenTo >= fnode.Len() {
		return
	}

	for *elects == value {
		prt := "===SimElectionsStart===\n\n"
		prt += "-------------------------\n"
		if fnode.Len() == 0 {
			return
		}

		//s := fnodes[*listenTo].State
		//eo := s.Elections.(*elections.Elections)

		prt = prt + "\n"
		for _, fn := range fnode.GetFnodes() {
			s := fn.State
			e := s.Elections.(*elections.Elections)
			if e.Adapter != nil {
				prt += e.Adapter.Status()
				prt += "\n"
				prt += e.Adapter.VolunteerControlsStatus()
				prt += "\n"
				//prt += e.Adapter.MessageLists()
				//prt += "\n"
			} else {
				prt += fmt.Sprintf("%s has no simelection\n", fn.State.GetFactomNodeName())
			}
		}

		prt = prt + "===SimElectionsEnd===\n"

		if prt != out {
			fmt.Println(prt)
			out = prt
		}

		time.Sleep(time.Second)
	}
}

func printElections(elects *int, value int, listenTo *int, wsapiNode *int) {
	out := ""

	if *listenTo < 0 || *listenTo >= fnode.Len() {
		return
	}

	for *elects == value {
		prt := "===ElectionsStart===\n\n"
		if fnode.Len() == 0 {
			return
		}

		s := fnode.Get(*listenTo).State
		eo := s.Elections.(*elections.Elections)

		prt = prt + fmt.Sprintf("%3s %15s %15s\n", "#", "Federated", "Audit")
		for i := 0; i < len(eo.Federated)+len(eo.Audit); i++ {
			fed := ""
			aud := ""
			if i < len(eo.Federated) {
				id := eo.Federated[i].GetChainID()
				f := lookup(id)
				if f != nil {
					fed = f.FactomNodeName
				}
			}
			if i < len(eo.Audit) {
				id := eo.Audit[i].GetChainID()
				a := lookup(id)
				if a != nil {
					aud = a.FactomNodeName
				}
			}
			if fed == "" && aud == "" {
				break
			}
			prt = prt + fmt.Sprintf("%3d %15s %15s\n", i, fed, aud)
		}

		prt = prt + "\n" + fnode.Get(0).State.Election0
		for i, _ := range eo.Federated {
			prt = prt + fmt.Sprintf("%4d ", i)
		}
		for i, _ := range eo.Audit {
			prt = prt + fmt.Sprintf("%4d ", i)
		}
		prt = prt + "\n"
		for _, fn := range fnode.GetFnodes() {
			s := fn.State
			if s.Elections.(*elections.Elections).Adapter != nil {
				e := s.Elections.(*elections.Elections).Electing
				prt += fmt.Sprintf("%2d ", e)
				if s.Elections.(*elections.Elections).Adapter.IsObserver() {
					prt += "O " // Observer
				} else {
					prt += "A " // Active
				}
			} else {
				prt += "__ _ " // Active
			}
			prt = prt + s.Election1 + s.Election2 + "\n"
		}

		prt = prt + "===ElectionsEnd===\n"

		if prt != out {
			fmt.Println(prt)
			out = prt
		}

		time.Sleep(time.Second)
	}
}

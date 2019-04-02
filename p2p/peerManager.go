package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var pmLogger = packageLogger.WithField("subpack", "peerManager")

// PeerManager is responsible for managing all the Peers, both online and offline
type PeerManager struct {
	controller *Controller
	config     *Configuration
	stop       chan interface{}
	Data       chan PeerParcel

	peerMutex  sync.RWMutex
	peerByHash PeerMap
	peerByIP   map[string]PeerMap
	//onlinePeers map[string]bool // set of online peers
	incoming uint
	outgoing uint

	specialIP map[string]bool

	lastPeerRequest time.Time

	logger *log.Entry
}

// NewPeerManager creates a new peer manager for the given controller
// configuration is shared between the two
func NewPeerManager(controller *Controller) *PeerManager {
	pm := &PeerManager{}
	pm.controller = controller
	pm.config = controller.Config

	pm.logger = pmLogger.WithFields(log.Fields{
		"node":    pm.config.NodeName,
		"port":    pm.config.ListenPort,
		"network": fmt.Sprintf("%#x", pm.config.Network)})
	pm.logger.WithField("peermanager_init", pm.controller.Config).Debugf("Initializing Peer Manager")

	pm.peerByHash = make(PeerMap)
	pm.peerByIP = make(map[string]PeerMap)

	pm.stop = make(chan interface{}, 1)
	pm.Data = make(chan PeerParcel, StandardChannelSize)

	// TODO parse config special peers

	return pm
}

// Start starts the peer manager
// reads from the seed and connect to peers
func (pm *PeerManager) Start() {
	pm.logger.Info("Starting the Peer Manager")

	// TODO discover from seed
	// 		parse and dial special peers
	//go pm.receiveData()
	go pm.managePeers()
	go pm.manageData()
}

// Stop shuts down the peer manager and all active connections
func (pm *PeerManager) Stop() {
	pm.stop <- true

	pm.peerMutex.RLock()
	defer pm.peerMutex.RUnlock()
	for _, p := range pm.peerByHash {
		p.GoOffline()
	}
}

func (pm *PeerManager) manageData() {
	for {
		data := <-pm.Data
		parcel := data.Parcel
		peer := data.Peer

		switch parcel.Header.Type {
		case TypeMessagePart: // deprecated
		case TypeHeartbeat: // deprecated
		case TypePing:
		case TypePong:
		case TypeAlert:

		case TypeMessage: // Application message, send it on.
			ApplicationMessagesReceived++
			BlockFreeChannelSend(pm.controller.FromNetwork, parcel)

		case TypePeerRequest:
			if time.Since(peer.lastPeerSend) >= pm.config.PeerRequestInterval {
				peer.lastPeerSend = time.Now()
				go pm.sharePeers(peer)
			} else {
				pm.logger.Warnf("Peer %s requested peer share sooner than expected", peer)
			}
		case TypePeerResponse:
			// TODO check here if we asked them for a peer request
			if time.Since(peer.lastPeerRequest) >= pm.config.PeerRequestInterval {
				peer.lastPeerRequest = time.Now()
				go pm.processPeers(peer, parcel)
			} else {
				pm.logger.Warnf("Peer %s sent us an umprompted peer share", peer)
			}
		default:
			pm.logger.Warnf("Peer %s sent unknown parcel.Header.Type?: %+v ", peer, parcel)
		}

	}

}

func (pm *PeerManager) processPeers(peer *Peer, parcel *Parcel) {
	for {
		select {
		case <-time.After(time.Second): // TODO make peerManageInterval variable
			for _, p := range pm.peerByHash {
				// request peers
				if time.Since(p.lastPeerRequest) > p.config.PeerRequestInterval {
					p.lastPeerRequest = time.Now()

					parcel := NewParcel(pm.config.Network, []byte("Peer Request"))
					parcel.Header.Type = TypePeerRequest
					p.Send(parcel)
				}
			}
		}
	}
}

func (pm *PeerManager) sharePeers(peer *Peer) {

}

func (pm *PeerManager) managePeers() {
	// remove old peers
	// search for duplicates

	/*	for {
		peer := <-pm.ShutDown
		pm.removePeer(peer)
		pm.Stop()
	}*/
}

func (pm *PeerManager) SpawnPeer(config *Configuration, address string, outgoing bool, listenPort string) *Peer {
	p := &Peer{Address: address, Outgoing: outgoing, state: Offline, ListenPort: listenPort}
	p.peerManager = pm
	p.logger = peerLogger.WithFields(log.Fields{
		"hash":       p.Hash,
		"address":    p.Address,
		"port":       p.Port,
		"listenPort": p.ListenPort,
		"outgoing":   p.Outgoing,
	})
	p.config = config
	p.stop = make(chan interface{}, 1)
	p.incoming = make(chan *Parcel, StandardChannelSize)
	p.Hash = address + ":" + listenPort // TODO make this a hash
	pm.addPeer(p)
	return p
}

// addPeer adds a peer to the manager system
func (pm *PeerManager) addPeer(peer *Peer) {
	pm.peerMutex.Lock()
	defer pm.peerMutex.Unlock()

	pm.peerByHash.Add(peer)
	if ip, ok := pm.peerByIP[peer.Address]; ok {
		ip.Add(peer)
	} else {
		pm.peerByIP[peer.Address] = PeerMap{peer.Hash: peer}
	}

	if peer.Outgoing {
		pm.outgoing++
	} else {
		pm.incoming++
	}
}

func (pm *PeerManager) removePeer(peer *Peer) {
	pm.peerMutex.Lock()
	defer pm.peerMutex.Unlock()
	pm.peerByHash.Remove(peer)
	delete(pm.peerByIP[peer.Address], peer.Hash)
}

func (pm *PeerManager) HandleIncoming(con net.Conn) {
	ip := con.RemoteAddr().String()
	special := pm.specialIP[ip]

	ipLog := pm.logger.WithField("remote_addr", ip)

	if !special {
		if pm.outgoing >= pm.config.Outgoing {
			ipLog.Info("Rejecting inbound connection because of inbound limit")
			con.Close()
			return
		} else if pm.config.RefuseIncoming || pm.config.RefuseUnknown {
			ipLog.WithFields(log.Fields{
				"RefuseIncoming": pm.config.RefuseIncoming,
				"RefuseUnknown":  pm.config.RefuseUnknown,
			}).Info("Rejecting inbound connection because of config settings")
			con.Close()
			return
		}
	}

	p := pm.SpawnPeer(pm.config, ip, false, "0")
	p.StartWithActiveConnection(con) // peer is online

	//c := NewConnection(con, pm.config)

	// TODO check if special
	// TODO check if incoming is maxed out
	// TODO add peer

	/*
		if ok, err := c.canConnectTo(conn); !ok {
			connLogger.Infof("Rejecting new connection request: %s", err)
			_ = conn.Close()
			continue
		}

		c.AddPeer(conn) // Sends command to add the peer to the peers list
		connLogger.Infof("Accepting new incoming connection")*/
}

func (pm *PeerManager) Broadcast(parcel *Parcel, full bool) {
	pm.peerMutex.RLock()
	defer pm.peerMutex.RUnlock()
	if full {
		for _, p := range pm.peerByHash {
			p.Send(parcel)
		}
		return
	}

	// fanout
	selection := pm.selectRandomPeers(pm.config.Fanout)
	for _, p := range selection {
		p.Send(parcel)
	}
	// TODO always send to special
}

func (pm *PeerManager) sortedOutgoing(desired int) []*Peer {
	var filtered []*Peer
	pm.peerMutex.RLock()
	for _, p := range pm.peerByHash {
		if p.IsOffline() && p.CanDial() && (!p.config.TrustedOnly || p.IsSpecial()) {
			filtered = append(filtered, p)
		}
	}
	pm.peerMutex.RUnlock()

	return filtered
}

func (pm *PeerManager) selectRandomPeers(count uint) []*Peer {
	pm.peerMutex.RLock()
	var peers []*Peer
	for _, p := range pm.peerByHash {
		if p.IsOnline() {
			peers = append(peers, p)
		}
	}
	pm.peerMutex.RUnlock() // unlock early before a shuffle

	// not enough to randomize
	if uint(len(peers)) <= count {
		return peers
	}

	shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	// TODO add special?
	return peers[:count]
}

func (pm *PeerManager) ToPeer(hash string, parcel *Parcel) {
	if hash == "" {
		if random := pm.selectRandomPeers(1); len(random) > 0 {
			random[0].Send(parcel)
		}
	} else {
		pm.peerMutex.RLock()
		defer pm.peerMutex.RUnlock()
		if peer, ok := pm.peerByHash[hash]; ok {
			peer.Send(parcel)
		}
	}
}

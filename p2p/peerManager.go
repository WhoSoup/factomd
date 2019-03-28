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
	config     *P2PConfiguration

	peerMutex   sync.RWMutex
	peerByHash  PeerMap
	peerByIP    map[string]PeerMap
	onlinePeers map[string]bool // set of online peers
	incoming    uint
	outgoing    uint

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

	// TODO parse config special peers

	return pm
}

// Start starts the peer manager
// reads from the seed and connect to peers
func (pm *PeerManager) Start() {
	pm.logger.Info("Starting the Peer Manager")

	// TODO discover from seed
	// 		parse and dial special peers
}

// Stop shuts down the peer manager and all active connections
func (pm *PeerManager) Stop() {

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

	p := NewPeer(pm.config, ip, false)
	p.HandleActiveConnection(con) // peer is online

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

func (pm *PeerManager) Broadcast(parcel Parcel, full bool) {
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

func (pm *PeerManager) selectRandomPeers(count uint) []*Peer {
	var peers []*Peer
	for i := range pm.onlinePeers {
		peers = append(peers, pm.peerByHash[i])
	}

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

func (pm *PeerManager) ToPeer(hash string, parcel Parcel) {
	if hash == "" {
		if random := pm.selectRandomPeers(1); len(random) > 0 {
			random[0].Send(parcel)
		}
	} else {
		if peer, ok := pm.peerByHash[hash]; ok {
			peer.Send(parcel)
		}
	}
}

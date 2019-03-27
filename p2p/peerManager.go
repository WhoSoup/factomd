package p2p

import (
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var pmLogger = packageLogger.WithField("subpack", "peerManager")

// PeerManager is responsible for managing all the Peers, both online and offline
type PeerManager struct {
	controller *Controller
	config     *P2PConfiguration

	peerByHash map[string]*Peer
	peerByIP   map[string][]*Peer

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

	pm.peerByHash = make(map[string]*Peer)
	pm.peerByIP = make(map[string][]*Peer)

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

func (pm *PeerManager) HandleIncoming(con net.Conn) {

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

}

func (pm *PeerManager) ToPeer(hash string, parcel Parcel) {

}

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var peerLogger = packageLogger.WithField("subpack", "peer")

// Data structures and functions related to peers (eg other nodes in the network)

type PeerType uint8

const (
	RegularPeer        PeerType = iota
	SpecialPeerConfig           // special peer defined in the config file
	SpecialPeerCmdLine          // special peer defined via the cmd line params
	PeerIncoming
	PeerOutgoing
)

// PeerState is the states for the Peer's state machine
type PeerState uint8

// The peer state machine's states
const (
	Offline PeerState = iota
	Connecting
	Online
)

type Peer struct {
	peerManager            *PeerManager
	conn                   *Connection
	state                  PeerState
	stateMutex             sync.RWMutex
	stop                   chan interface{}
	Outgoing               bool
	config                 *Configuration
	lastPeerRequest        time.Time
	lastPeerSend           time.Time
	incoming               chan *Parcel
	connectionAttempt      time.Time
	connectionAttemptCount uint

	ListenPort string

	QualityScore int32     // 0 is neutral quality, negative is a bad peer.
	Address      string    // Must be in form of x.x.x.x
	Port         string    // Must be in form of xxxx
	NodeID       uint64    // a nonce to distinguish multiple nodes behind one IP address
	Hash         string    // This is more of a connection ID than hash right now.
	Location     uint32    // IP address as an int.
	Network      NetworkID // The network this peer reference lives on.
	Type         PeerType
	Connections  int                  // Number of successful connections.
	LastContact  time.Time            // Keep track of how long ago we talked to the peer.
	Source       map[string]time.Time // source where we heard from the peer.

	// logging
	logger *log.Entry
}

func (p *Peer) String() string {
	return fmt.Sprintf("%s %s:%s", p.Hash, p.Address, p.ListenPort)
}

func (p *Peer) StartToDial() {
	if !p.CanDial() {
		p.logger.Errorf("Attempted to connect to a peer with no remote listen port")
		return
	}

	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if p.conn != nil {
		p.logger.WithField("old_conn", p.conn).Warn("Peer started to dial despite not being offline")
		p.conn.Stop()
		p.conn = nil
	}

	p.Outgoing = true
	p.state = Connecting
	p.connectionAttempt = time.Now()
	p.connectionAttemptCount++
	remote := fmt.Sprintf("%s:%s", p.Address, p.ListenPort)
	con, err := net.Dial("tcp", remote)
	if err != nil {
		p.logger.WithError(err).Infof("Unable to connect to peer")
		return
	}

	p.startInternal(con)
}

func (p *Peer) StartWithActiveConnection(con net.Conn) {
	if p.conn != nil {
		p.logger.WithField("old_conn", p.conn).Warn("Peer given new connection despite having old one")
		p.conn.Stop()
		p.conn = nil
	}
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.startInternal(con)
}

// startInternal is the common functionality for both dialing and accepting a connection
// is under locked mutex from superior function
func (p *Peer) startInternal(con net.Conn) {
	p.conn = NewConnection(con, p.config, p.incoming)
	p.conn.Start()
	p.state = Online
	go p.monitorConnection() // this will die when the connection is closed
}

func (p *Peer) GoOffline() {
	p.stateMutex.Lock()
	defer p.stateMutex.Lock()
	p.state = Offline
	p.stop <- true
}

func (p *Peer) Send(parcel *Parcel) {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()
	if p.state != Online {
		if p.state == Connecting {
			log.Error("Tried to send parcel connection still connecting")
		} else {
			log.Error("Tried to send parcel on offline connection")
		}
		return
	}
	// TODO check peer state machine
	parcel.Header.NodeID = p.config.NodeID
	parcel.Header.PeerPort = string(p.config.ListenPort) // notify other side of our port
	BlockFreeParcelSend(p.conn.Outgoing, parcel)
}

func (p *Peer) CanDial() bool {
	return p.ListenPort != "0"
}

func (p *Peer) IsOnline() bool {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()
	return p.state == Online
}
func (p *Peer) IsOffline() bool {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()
	return p.state == Offline
}

// monitorConnection watches the underlying Connection.
// Any data arriving via connection will be passed on to the peer manager.
// If the connection dies, change state to Offline
func (p *Peer) monitorConnection() {
	for {
		select {
		case <-p.stop: // manual stop, we need to tear down connection
			p.conn.Stop()
			p.conn = nil
			p.connectionAttemptCount = 0
			return
		case <-p.conn.Errors:
			p.stateMutex.Lock()
			defer p.stateMutex.Lock()
			p.state = Offline
			p.conn = nil // if an error arrives here, the connection already stops itself
			p.connectionAttemptCount = 0
			return
		case parcel := <-p.incoming:
			if newport := parcel.Header.PeerPort; newport != p.ListenPort {
				p.logger.WithFields(log.Fields{"old": p.ListenPort, "new": newport}).Debugf("Listen port changed")
				p.ListenPort = newport
			}
			if nodeid := parcel.Header.NodeID; nodeid != p.NodeID {
				p.logger.WithFields(log.Fields{"old": p.NodeID, "new": nodeid}).Debugf("NodeID changed")
				p.NodeID = nodeid
			}
			p.peerManager.Data <- PeerParcel{Peer: p, Parcel: parcel} // TODO this is potentially blocking
		}
	}
}

func (p *Peer) Init(address string, port string, quality int32, peerType PeerType, connections int) *Peer {

	p.logger = peerLogger.WithFields(log.Fields{
		"address":  address,
		"port":     port,
		"peerType": peerType,
	})
	if net.ParseIP(address) == nil {
		ipAddress, err := net.LookupHost(address)
		if err != nil {
			p.logger.Errorf("Init: LookupHost(%v) failed. %v ", address, err)
			// is there a way to abandon this peer at this point? -- clay
		} else {
			address = ipAddress[0]
		}
	}

	p.Address = address
	p.Port = port
	p.QualityScore = quality
	p.generatePeerHash()
	p.Type = peerType
	p.Location = p.LocationFromAddress()
	p.Source = map[string]time.Time{}
	p.Network = CurrentNetwork
	return p
}

func (p *Peer) generatePeerHash() {
	p.Hash = fmt.Sprintf("%s:%s %x", p.Address, p.Port, rand.Int63())
}

func (p *Peer) AddressPort() string {
	return p.Address + ":" + p.Port
}

func (p *Peer) PeerIdent() string {
	return p.Hash[0:12] + "-" + p.Address + ":" + p.Port
}

func (p *Peer) PeerFixedIdent() string {
	address := fmt.Sprintf("%16s", p.Address)
	return p.Hash[0:12] + "-" + address + ":" + p.Port
}

func (p *Peer) PeerLogFields() log.Fields {
	return log.Fields{
		"address":   p.Address,
		"port":      p.Port,
		"peer_type": p.PeerTypeString(),
	}
}

// gets the last source where this peer was seen
func (p *Peer) LastSource() (result string) {
	var maxTime time.Time

	for source, lastSeen := range p.Source {
		if lastSeen.After(maxTime) {
			maxTime = lastSeen
			result = source
		}
	}

	return
}

// TODO Hadn't considered IPV6 address support.
// TODO Need to audit all the net code to check IPv6 addresses
// Here's an IPv6 conversion:
// Ref: http://stackoverflow.com/questions/23297141/golang-net-ip-to-ipv6-from-mysql-as-decimal39-0-conversion
// func ipv6ToInt(IPv6Addr net.IP) *big.Int {
//     IPv6Int := big.NewInt(0)
//     IPv6Int.SetBytes(IPv6Addr)
//     return IPv6Int
// }
// Problem is we're working with string addresses, may never have made a connection.
// TODO - we might have a DNS address, not iP address and need to resolve it!
// locationFromAddress converts the peers address into a uint32 "location" numeric
func (p *Peer) LocationFromAddress() (location uint32) {
	location = 0
	// Split the IPv4 octets
	ip := net.ParseIP(p.Address)
	if ip == nil {
		ipAddress, err := net.LookupHost(p.Address)
		if err != nil {
			p.logger.Debugf("LocationFromAddress(%v) failed. %v ", p.Address, err)
			p.logger.Debugf("Invalid Peer Address: %v", p.Address)
			p.logger.Debugf("Peer: %s has Location: %d", p.Hash, location)
			return 0 // We use location on 0 to say invalid
		}
		p.Address = ipAddress[0]
		ip = net.ParseIP(p.Address)
	}
	if len(ip) == 16 { // If we got back an IP6 (16 byte) address, use the last 4 byte
		ip = ip[12:]
	}
	// Turn into uint32
	location += uint32(ip[0]) << 24
	location += uint32(ip[1]) << 16
	location += uint32(ip[2]) << 8
	location += uint32(ip[3])
	p.logger.Debugf("Peer: %s has Location: %d", p.Hash, location)
	return location
}

func (p *Peer) IsSamePeerAs(netAddress net.Addr) bool {
	address, _, err := net.SplitHostPort(netAddress.String())
	if err != nil {
		return false
	}
	return address == p.Address
}

// merit increases a peers reputation
func (p *Peer) merit() {
	if 2147483000 > p.QualityScore {
		p.QualityScore++
	}
}

// demerit decreases a peers reputation
func (p *Peer) demerit() {
	if -2147483000 < p.QualityScore {
		//p.QualityScore--
	}
}

func (p *Peer) IsSpecial() bool {
	return p.Type == SpecialPeerConfig || p.Type == SpecialPeerCmdLine
}

func (p *Peer) PeerTypeString() string {
	switch p.Type {
	case RegularPeer:
		return "regular"
	case SpecialPeerConfig:
		return "special_config"
	case SpecialPeerCmdLine:
		return "special_cmdline"
	default:
		return "unknown"
	}
}

// sort.Sort interface implementation
type PeerQualitySort []Peer

func (p PeerQualitySort) Len() int {
	return len(p)
}
func (p PeerQualitySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PeerQualitySort) Less(i, j int) bool {
	return p[i].QualityScore < p[j].QualityScore
}

// sort.Sort interface implementation
type PeerDistanceSort []*Peer

func (p PeerDistanceSort) Len() int {
	return len(p)
}
func (p PeerDistanceSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PeerDistanceSort) Less(i, j int) bool {
	return p[i].Location < p[j].Location
}

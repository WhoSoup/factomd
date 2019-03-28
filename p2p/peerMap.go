package p2p

type PeerMap map[string]*Peer

func (pm PeerMap) Add(p *Peer) {
	pm[p.Hash] = p
}

func (pm PeerMap) Remove(p *Peer) {
	delete(pm, p.Hash)
}

func (pm PeerMap) RemoveHash(hash string) {
	delete(pm, hash)
}

package p2p

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// LimitedListener will block multiple connection attempts from a single ip
// within a specific timeframe
type LimitedListener struct {
	net.Listener

	limit          time.Duration
	lastConnection time.Time
	history        []limitedConnect
}

type limitedConnect struct {
	address string
	time    time.Time
}

func NewLimitedListener(address string, limit time.Duration) (*LimitedListener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	if limit < 0 {
		return nil, fmt.Errorf("Invalid time limit (negative)")
	}
	return &LimitedListener{
		Listener:       l,
		limit:          limit,
		lastConnection: time.Time{},
		history:        nil,
	}, nil
}

// clearHistory truncates the history to only relevant entries
func (ll *LimitedListener) clearHistory() {
	tl := time.Now().Add(-ll.limit) // get timelimit of range to check

	// no connection made in the last X seconds
	// the vast majority of connections will proc this
	if ll.lastConnection.Before(tl) {
		ll.history = nil // reset and release to gc
	}

	if len(ll.history) > 0 {
		i := 0
		for ; i < len(ll.history); i++ {
			if ll.history[i].time.After(tl) { // inside target range
				break
			}
		}

		if i >= len(ll.history) {
			ll.history = nil
		} else {
			ll.history = ll.history[i:]
		}
	}
}

// isInHistory checks if an address has connected in the last X seconds
// clears history before checking
func (ll *LimitedListener) isInHistory(addr string) bool {
	ll.clearHistory()

	for _, h := range ll.history {
		if h.address == addr {
			return true
		}
	}
	return false
}

// addToHistory adds an address to the system at the current time
func (ll *LimitedListener) addToHistory(addr string) {
	ll.history = append(ll.history, limitedConnect{address: addr, time: time.Now()})
	ll.lastConnection = time.Now()
}

// Accept accepts a connection if no other connection attempt from that ip has been made
// in the specified time frame
func (ll *LimitedListener) Accept() (net.Conn, error) {
	con, err := ll.Listener.Accept()
	if err != nil {
		return nil, err
	}

	addr := strings.Split(con.RemoteAddr().String(), ":")
	if ll.isInHistory(addr[0]) {
		con.Close()
		return nil, fmt.Errorf("connection rate limit exceeded")
	}

	ll.addToHistory(addr[0])
	return con, nil
}

// limitListenerSources will limit the number of connections allowed to 1
// connection per ip per second. Any more than that it will reject.
// The rate limiting is pretty dumb, when a connection is made, no other
// connection by that IP is allowed for 1 second
type limitListenerSources struct {
	net.Listener

	// This is not garbage collected, be aware it is a small memory leak
	// TODO: Fix the memory leak issue.
	// Technically it is bounded by the number of possible IP addresses, so it will consume
	// at most ~34.4GB if every possible ipv4 address connects.
	accepted map[string]time.Time
}

// Accept is overridden here to the default Accept
func (l *limitListenerSources) Accept() (net.Conn, error) {
	// We need to accept the connection first to determine the IP address.
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Grab the address, check for last connection
	addr := strings.Split(c.RemoteAddr().String(), ":")
	if v, ok := l.accepted[addr[0]]; !ok || time.Since(v) > time.Second {
		l.accepted[addr[0]] = time.Now()
		return c, nil
	}
	c.Close()
	return nil, fmt.Errorf("rate limited")
}

func LimitListenerSources(l net.Listener) net.Listener {
	return &limitListenerSources{Listener: l, accepted: make(map[string]time.Time)}
}

// limitListenerAll will limit the number of connections allowed to 1
// connection per second. Any more than that it will reject.
// The rate limiting is pretty dumb, when a connection is made, no other
// connection is allowed for 1 second
type limitListenerAll struct {
	net.Listener

	last time.Time
}

// Accept is overridden here to the default Accept
func (l *limitListenerAll) Accept() (net.Conn, error) {
	if time.Since(l.last) < time.Second {
		return nil, fmt.Errorf("rate limited")
	}

	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	l.last = time.Now()
	return c, nil
}

func LimitListenerAll(l net.Listener) net.Listener {
	return &limitListenerAll{Listener: l}
}

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
	accepted       map[string]time.Time
}

func LimitedListen(network, address string, limit time.Duration) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return LimitedListener{
		Listener:       l,
		limit:          limit,
		lastConnection: time.Time{},
		accepted:       make(map[string]time.Time),
	}, nil
}

func (ll LimitedListener) Accept() (net.Conn, error) {
	con, err := ll.Listener.Accept()
	if err != nil {
		return nil, err
	}

	addr := strings.Split(con.RemoteAddr().String(), ":")
	if t, ok := ll.accepted[addr[0]]; ok && time.Since(t) > ll.limit {
		ll.accepted[addr[0]] = time.Now()
		con.Close()
		return nil, fmt.Errorf("connection rate limit exceeded")
	}

	// if no connection has been made in a while, reset the map
	if len(ll.accepted) > 16 && time.Since(ll.lastConnection) > ll.limit {
		ll.accepted = make(map[string]time.Time)
	}

	ll.lastConnection = time.Now()
	ll.accepted[addr[0]] = time.Now()
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
	fmt.Println("Accepted")

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

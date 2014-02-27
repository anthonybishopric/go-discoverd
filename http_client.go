package discoverd

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
)

// HTTPClient returns a HTTP client configured to use discoverd to lookup hostnames.
func (c *Client) HTTPClient() *http.Client {
	return &http.Client{Transport: &http.Transport{Dial: c.DialFunc(nil)}}
}

type DialFunc func(network, addr string) (net.Conn, error)

// DialFunc returns a DialFunc that uses discoverd to lookup hostnames. If f is
// provided, it used to Dial after looking up an address.
func (c *Client) DialFunc(f DialFunc) DialFunc {
	return newDialer(c, f).Dial
}

func newDialer(c *Client, f DialFunc) *dialer {
	d := &dialer{c: c, sets: make(map[string]ServiceSet), dial: f}
	if d.dial == nil {
		d.dial = net.Dial
	}
	return d
}

type dialer struct {
	c    *Client
	dial DialFunc
	sets map[string]ServiceSet
	mtx  sync.RWMutex
}

var ErrNoServices = errors.New("discoverd: no online services found")

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	name := strings.SplitN(addr, ":", 2)[0]
	set, err := d.getSet(name)
	if err != nil {
		return nil, err
	}

	addrs := set.Addrs()
	if len(addrs) == 0 {
		return nil, ErrNoServices
	}

	return d.dial(network, addrs[rand.Intn(len(addrs))])
}

func (d *dialer) getSet(name string) (ServiceSet, error) {
	d.mtx.RLock()
	set := d.sets[name]
	d.mtx.RUnlock()
	if set == nil {
		return d.createSet(name)
	}
	return set, nil
}

func (d *dialer) createSet(name string) (ServiceSet, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if set, ok := d.sets[name]; ok {
		return set, nil
	}

	set, err := d.c.NewServiceSet(name)
	if err != nil {
		return nil, err
	}
	d.sets[name] = set
	return set, nil
}

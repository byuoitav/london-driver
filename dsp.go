/*
Package london provides a struct for controlling BSS London DSPs.

Supported Devices

This is a list of devices that BYU currently uses in production, controlled with this driver.

	BSS BLU-50 https://bssaudio.com/en/products/blu-50
	BSS BLU-100 https://bssaudio.com/en/products/blu-100

This list is not comprehensive.
*/
package london

import (
	"context"
	"net"

	"github.com/byuoitav/connpool"
)

// DSP represents a DSP being controlled.
type DSP struct {
	address string
	pool    *connpool.Pool

	logger Logger
}

// New returns a new DSP with the given address.
func New(addr string, opts ...Option) *DSP {
	options := options{
		ttl:   _defaultTTL,
		delay: _defaultDelay,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	d := &DSP{
		address: addr,
		pool: &connpool.Pool{
			TTL:   options.ttl,
			Delay: options.delay,
		},
		logger: options.logger,
	}

	d.pool.NewConnection = func(ctx context.Context) (net.Conn, error) {
		dial := net.Dialer{}
		return dial.DialContext(ctx, "tcp", d.address+":1023")
	}

	return d
}

func (d *DSP) GetInfo(ctx context.Context) (interface{}, error) {
	return nil, nil
}

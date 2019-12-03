package london

import (
	"context"
	"net"
	"time"

	"github.com/byuoitav/connpool"
)

type DSP struct {
	Address string
	pool    *connpool.Pool
}

var (
	_defaultTTL   = 45 * time.Second
	_defaultDelay = 400 * time.Millisecond
)

type options struct {
	ttl   time.Duration
	delay time.Duration
}

type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func WithTTL(t time.Duration) Option {
	return optionFunc(func(o *options) {
		o.ttl = t
	})
}

func WithDelay(t time.Duration) Option {
	return optionFunc(func(o *options) {
		o.delay = t
	})
}

func NewDSP(addr string, opts ...Option) *DSP {
	options := options{
		ttl:   _defaultTTL,
		delay: _defaultDelay,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	d := &DSP{
		Address: addr,
		pool: &connpool.Pool{
			TTL:   options.ttl,
			Delay: options.delay,
		},
	}

	d.pool.NewConnection = func(ctx context.Context) (net.Conn, error) {
		dial := net.Dialer{}
		return dial.DialContext(ctx, "tcp", d.Address+":1023")
	}

	return d
}

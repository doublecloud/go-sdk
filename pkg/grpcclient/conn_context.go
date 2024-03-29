package grpcclient

import (
	"context"
	"errors"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
)

var ErrConnContextClosed = errors.New("grpcclient: client connection context closed")

type DialError struct {
	Err error
	Add string
}

func (d *DialError) Error() string {
	return "error dialing endpoint '" + d.Add + "': " + d.Err.Error()
}

//go:generate mockery -name=ConnContext

type ConnContext interface {
	GetConn(ctx context.Context, addr string) (*grpc.ClientConn, error)
	CallOptions() []grpc.CallOption
	Shutdown(context.Context) error
}

type LazyConnContextOption func(*lazyConnContextOptions)

type lazyConnContextOptions struct {
	dialOpts []grpc.DialOption
	callOpts []grpc.CallOption
}

func DialOptions(dopts ...grpc.DialOption) LazyConnContextOption {
	return func(o *lazyConnContextOptions) {
		o.dialOpts = append(o.dialOpts, dopts...)
	}
}

func CallOptions(copts ...grpc.CallOption) LazyConnContextOption {
	return func(o *lazyConnContextOptions) {
		o.callOpts = append(o.callOpts, copts...)
	}
}

type lazyConnContext struct {
	opts *lazyConnContextOptions

	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	conns   map[string]*grpc.ClientConn
	closed  bool
	closing bool

	dial     singleflight.Group
	shutdown singleflight.Group
}

func NewLazyConnContext(opt ...LazyConnContextOption) ConnContext {
	opts := &lazyConnContextOptions{}
	for _, o := range opt {
		o(opts)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &lazyConnContext{
		opts:   opts,
		ctx:    ctx,
		cancel: cancel,
		conns:  map[string]*grpc.ClientConn{},
	}
}

func (cc *lazyConnContext) GetConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	cc.mu.Lock()
	if cc.closed || cc.closing {
		cc.mu.Unlock()
		return nil, ErrConnContextClosed
	}
	if conn, ok := cc.conns[addr]; ok {
		cc.mu.Unlock()
		return conn, nil
	}
	cc.mu.Unlock()

	result, err, _ := cc.dial.Do(addr, func() (any, error) {
		conn, err := grpc.DialContext(cc.ctx, addr, cc.opts.dialOpts...)
		if err != nil {
			if err == cc.ctx.Err() {
				err = ErrConnContextClosed
			} else {
				err = &DialError{err, addr}
			}
			// return connAndErr{err: err}
			return nil, err
		}
		defer cc.mu.Unlock()
		cc.mu.Lock()

		if cc.closed || cc.closing {
			// cc.mu.Unlock()
			// we swallow error here, since the client doesn't care about it
			_ = conn.Close()
			// return connAndErr{conn: nil, err: ErrConnContextClosed}
			return nil, ErrConnContextClosed
		}
		cc.conns[addr] = conn
		// cc.mu.Unlock()
		// return connAndErr{conn: conn}
		return conn, nil
	})
	ce := result.(*grpc.ClientConn)
	// return ce.conn, ce.err
	return ce, err
}

func (cc *lazyConnContext) CallOptions() []grpc.CallOption {
	callOpts := make([]grpc.CallOption, len(cc.opts.callOpts))
	copy(callOpts, cc.opts.callOpts)
	return callOpts
}

func (cc *lazyConnContext) Shutdown(ctx context.Context) error {
	cc.mu.Lock()
	if cc.closed {
		cc.mu.Unlock()
		return nil
	}
	cc.closing = true
	cc.mu.Unlock()

	// result := cc.shutdown.Do(func() interface{} {
	result, err, _ := cc.shutdown.Do("shutdown", func() (any, error) {
		cc.mu.Lock()
		cc.cancel()
		conns := make([]*grpc.ClientConn, 0, len(cc.conns))
		for _, conn := range cc.conns {
			conns = append(conns, conn)
		}
		cc.mu.Unlock()

		var errs error
		for _, conn := range conns {
			err := conn.Close()
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		cc.mu.Lock()
		cc.closed = true
		cc.closing = false
		cc.mu.Unlock()
		return nil, errs
	})

	if result == nil {
		return nil
	}
	return err
}

type connAndErr struct {
	conn *grpc.ClientConn
	err  error
}

package sockstun

import (
	"context"
	"fmt"
	"io"
	"log" //nolint:depguard // TODO: Replace by log/slog
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
	"golang.org/x/sync/errgroup"
)

type fwdRule struct {
	name                  string
	localSock, remoteSock string
}

func (r fwdRule) String() string {
	return fmt.Sprintf("%s (%s->%s)", r.name, r.localSock, r.remoteSock)
}

type SOCKSTunnel struct {
	proto       string
	socksDialer proxy.ContextDialer
	rwTimeout   time.Duration
	fwdTable    []fwdRule
	fwdTableMu  sync.RWMutex
	log         *log.Logger
}

func (st *SOCKSTunnel) Add(name, lsock, rsock string) {
	r := fwdRule{name: name, localSock: lsock, remoteSock: rsock}
	st.fwdTableMu.Lock()
	st.fwdTable = append(st.fwdTable, r)
	st.fwdTableMu.Unlock()
}

func (st *SOCKSTunnel) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	st.fwdTableMu.RLock()
	for _, r := range st.fwdTable {
		eg.Go(func() error {
			return st.enable(ctx, r)
		})
	}
	st.fwdTableMu.RUnlock()
	return eg.Wait() //nolint:wrapcheck // No need to wrap this
}

func (st *SOCKSTunnel) enable(ctx context.Context, r fwdRule) error {
	l, err := net.Listen(st.proto, r.localSock)
	if err != nil {
		return errors.Wrap(err, "failed to listen on local socket")
	}
	shutdown := (func() func() {
		var once sync.Once
		return func() {
			once.Do(func() {
				if err := l.Close(); err != nil {
					st.log.Printf("failed to close local socket %s: %v", r.localSock, err)
				}
			})
		}
	})()
	defer shutdown()
	st.log.Printf("enabling proxy rule %s", r)

	go func() {
		<-ctx.Done()
		shutdown()
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			// TODO(leon): remove this string search. That might involve
			// modifying the standard library to return better error types.
			if strings.Contains(err.Error(), "use of closed network connection") {
				return ctx.Err() //nolint:wrapcheck // No need to wrap this
			}
			st.log.Printf("failed to accept on local socket %s: %v", r.localSock, err)
			continue
		}
		go func() {
			if err := st.handle(ctx, conn, r); err != nil {
				st.log.Printf("failed to handle conn on local socket %s: %v", r.localSock, err)
			}
		}()
	}
}

func (st *SOCKSTunnel) handle(ctx context.Context, conn net.Conn, r fwdRule) error {
	var (
		doCleanup   = true
		cleanupFunc = func() {
			if err := conn.Close(); err != nil {
				st.log.Printf("%s: failed to close: %v", r, err)
			}
		}
		cleanup = (func() func() {
			var once sync.Once
			return func() {
				once.Do(cleanupFunc)
			}
		})()
	)
	defer func() {
		if doCleanup {
			cleanup()
		}
	}()

	sconn, err := st.socksDialer.DialContext(ctx, st.proto, r.remoteSock)
	if err != nil {
		return errors.Wrap(err, "failed to dial SOCKS proxy")
	}
	oldFunc := cleanupFunc
	cleanupFunc = func() {
		oldFunc()
		if err := sconn.Close(); err != nil {
			st.log.Printf("%s: failed to close SOCKS conn: %v", r, err)
		}
	}

	if t := st.rwTimeout; t > 0 {
		dl := time.Now().Add(t)
		if err := conn.SetDeadline(dl); err != nil {
			return errors.Wrap(err, "failed to set conn deadline")
		}
		if err := sconn.SetDeadline(dl); err != nil {
			return errors.Wrap(err, "failed to set SOCKS deadline")
		}
	}

	pipe := func(src, dst net.Conn) {
		defer cleanup()

		const maxSize = 65535
		buf := make([]byte, maxSize)
		for {
			n, err := src.Read(buf)
			if err != nil {
				// TODO(leon): remove this string search. That might involve
				// modifying the standard library to return better error types.
				if !errors.Is(err, io.EOF) && !strings.Contains(err.Error(), "use of closed network connection") {
					st.log.Printf("%s: failed to read: %v", r, err)
				}
				return
			}
			b := buf[:n]

			_, err = dst.Write(b)
			if err != nil {
				st.log.Printf("%s: failed to write: %v", r, err)
				return
			}
		}
	}

	doCleanup = false
	go pipe(conn, sconn)
	go pipe(sconn, conn)
	return nil
}

func New(socksURI string, rwTimeout time.Duration, logger *log.Logger) (*SOCKSTunnel, error) {
	su, err := url.Parse(socksURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse SOCKS URL")
	}
	dialer, err := proxy.FromURL(su, proxy.Direct)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create SOCKS dialer")
	}

	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		// This will never happen. proxy.Direct implements proxy.ContextDialer.
		panic("failed to type assert to proxy.ContextDialer") //nolint:forbidigo // Somewhat OK here
	}

	return &SOCKSTunnel{
		proto:       "tcp",
		socksDialer: ctxDialer,
		rwTimeout:   rwTimeout,
		log:         logger,
	}, nil
}

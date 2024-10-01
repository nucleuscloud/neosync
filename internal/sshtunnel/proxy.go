package sshtunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type NetProxy struct {
	source      *Endpoint
	destination *Endpoint
	dailer      Dialer

	isRunning atomic.Bool
	close     chan any

	shutdowns *sync.Map
}

// Create a new proxy that when started will bind to the source endpoint
func NewProxy(
	source *Endpoint,
	destination *Endpoint,
	dialer Dialer,
) *NetProxy {
	return &NetProxy{
		source:      source,
		destination: destination,
		dailer:      dialer,
	}
}

// The proxy binds to localhost on a random port
func NewHostProxy(
	destination *Endpoint,
	dialer Dialer,
) *NetProxy {
	return &NetProxy{
		source:      NewEndpoint("localhost", 0),
		destination: destination,
		dailer:      dialer,
	}
}

// I think this thing was supposed to act as a traditional proxy.
// It takes in a source, destination, and a tunnel dialer.
// When the proxy is started, it will start a TCP server
// and forward incoming traffic to the destination
// I think the source is useful if you want to bind to a specific local port
// Port 0 is the special case that allows you to bind to any port

func (n *NetProxy) GetSourceValues() (host string, port int) {
	return n.source.GetValues()
}

func (n *NetProxy) Start(ctx context.Context, proxyerrs chan<- error, logger *slog.Logger) error {
	listener, err := net.Listen("tcp", n.source.String())
	if err != nil {
		return fmt.Errorf("unable to start net proxy listener: %w", err)
	}

	go n.proxy(ctx, listener, logger)
	return nil
}

func (n *NetProxy) proxy(ctx context.Context, listener net.Listener, logger *slog.Logger) {
	defer func() {
		if err := listener.Close(); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logger.Error("failed to close tunnel listener", "error", err)
			}
		}
	}()

	n.isRunning.Store(true)

	for n.isRunning.Load() {
		if ctx.Err() != nil {
			logger.Debug("proxy ctx in error state, exiting serve loop")
			return
		}
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Debug("listener closed, stopping serve loop")
				return
			}
			logger.Error("failed to accept proxy connection", "error", err)
			continue
		}

		logger.Debug("accepted new proxy connection", "remoteAddr", conn.RemoteAddr().String())
		sessionId := uuid.NewString()
		shutdown := make(chan any)
		n.shutdowns.Store(sessionId, shutdown)

		go func() {
			defer func() {
				n.shutdowns.Delete(sessionId)
				if err := conn.Close(); err != nil {
					if !errors.Is(err, net.ErrClosed) {
						logger.Error("failed to close tunnel connection for session", "error", err, "sessionId", sessionId)
					}
				}
			}()
			select {
			case <-n.close:
				logger.Debug("received close signal, closing connection", "sessionId", sessionId)
			case <-shutdown:
				logger.Debug("received shutdown signal for session", "sessionId", sessionId)
			default:
				err := n.handleConnection(ctx, conn, sessionId, shutdown, logger.With("sessionId", sessionId))
				if err != nil {
					// todo: do something with this error
					logger.Debug(err.Error())
				}
			}
		}()
	}
}

func (n *NetProxy) handleConnection(
	ctx context.Context,
	conn net.Conn,
	sessionId string,
	shutdown <-chan any,
	logger *slog.Logger,
) error {
	destinationConn, err := n.dailer.DialContext(ctx, "tcp", n.destination.String())
	if err != nil {
		return fmt.Errorf("unable to dial destination connection: %w", err)
	}
	logger.Debug(fmt.Sprintf("successfully connected to %s", n.destination.String()))

	done := make(chan error, 2)
	go func() {
		select {
		case <-ctx.Done():
			logger.Debug("received ctx done while proxying connection, issuing shutdown")
			conn.Close()
			destinationConn.Close()
			logger.Debug("shutdown complete")
		case <-shutdown:
			logger.Debug("shutdown of proxy tunnel was issued")
			conn.Close()
			destinationConn.Close()
			logger.Debug("shutdown complete")
		case <-done:
			logger.Debug("received done while proxying, shutting down connections")
			n.shutdowns.Delete(sessionId)
			conn.Close()
			destinationConn.Close()
			logger.Debug("shutdown complete")
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		done <- proxyConnection(conn, destinationConn, logger.With("direction", "remote->local"))
	}()
	go func() {
		defer wg.Done()
		done <- proxyConnection(destinationConn, conn, logger.With("direction", "local->remote"))
	}()
	wg.Wait()
	logger.Debug("proxy connection session complete")
	return nil
}

func proxyConnection(writer, reader net.Conn, logger *slog.Logger) error {
	_, err := io.Copy(writer, reader)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			// This can be a common error if the thing using the ssh connection was abruptly closed.
			// This is common if a user is trying to test their database connection, but they've given Neosync bad credentials
			// or something else that causes the server to force close the client connection
			logger.Warn("connection was closed before reaching end of input", "error", err)
		} else {
			logger.Error("unexpected error while streaming through tunnel", "error", err)
		}
	} else {
		logger.Debug("ssh tunnel stream completed successfully")
	}
	return err
}

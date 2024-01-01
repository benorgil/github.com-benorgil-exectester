package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	// This is a somewhat popular library (datadog uses it)
	// but it isn't currently being developed and has some quirks
	"github.com/cenkalti/backoff/v4"
)

// Unix socket client
type unixSocket struct {
	socketName string
	// The socket conn timeout. Probably should not even be configurable
	// because its only for the dial command. Retries with backoff are
	// used to retry and reconnect instead.
	timeout    int
	logger     slog.Logger
	dialer     net.Dialer
	conn       net.Conn
	context    context.Context
	cancelFunc func()
	connClose  func() error
}

// Wrap the tear down methods in a single func
func (s *unixSocket) close() {
	s.cancelFunc()
	s.connClose()
}

// Retries connection to socket and updates object state with new connection
func (s *unixSocket) connectTounixSocket() error {
	s.dialer = net.Dialer{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
	s.dialer.LocalAddr = nil
	addr := net.UnixAddr{Name: s.socketName, Net: "unix"}

	connect := func() (net.Conn, error) {
		conn, err := s.dialer.DialContext(ctx, "unix", addr.String())
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to connect to '%v'. Error: '%v'. Retrying...", s.socketName, err))
			cancel()
			// Reset context before retrying
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
			return nil, err
		}
		return conn, nil
	}

	// Using all defaults and just setting a timeout via MaxElapsedTime
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Second
	b.MaxElapsedTime = time.Duration(s.timeout) * time.Second

	conn, err := backoff.RetryWithData(connect, b)

	if err != nil {
		s.logger.Error(fmt.Sprintf("Timed out connecting to '%v'", s.socketName))
		return err
	}

	s.conn = conn
	s.context = ctx
	s.cancelFunc = cancel
	s.connClose = conn.Close

	return nil
}

func getunixSocket(logger slog.Logger, socketName string, timeout int) (*unixSocket, error) {
	// Initial required params
	s := &unixSocket{
		socketName: socketName,
		timeout:    timeout,
		logger:     logger,
	}
	// connectTounixSocket() requires the above params
	err := s.connectTounixSocket()
	return s, err
}

func (s *unixSocket) sendTounixSocket(data string) error {
	_, err := s.conn.Write([]byte(data))
	return err
}

// WARNING: Reading from the socket buffer consumes those messages. It will be competing with
// anything else thats connected to the socket
func (s *unixSocket) readFromunixSocket(logger slog.Logger, exitMsg string) (response string, err error) {
	for {
		buf := make([]byte, 1024)
		n, err := s.conn.Read(buf)
		response = string(buf[0:n])

		if err != nil {
			// Check err to see if socket was closed
			if err.Error() == "EOF" {
				logger.Error(fmt.Sprintf("'%v' returned 'EOF'", s.socketName))
			} else {
				logger.Error(fmt.Sprintf("Unknown error from '%v': '%v'", s.socketName, err.Error()))
			}

			// Attempt to reconnect
			logger.Error(fmt.Sprintf("Retrying connection to '%v'", s.socketName))
			err := s.connectTounixSocket()
			if err != nil {
				uerr := fmt.Errorf(fmt.Sprintf(
					"Reached timeout of '%v' waiting for socket '%v'. "+
						"Error from socket: %v",
					s.timeout, s.socketName, err.Error()))
				return "", uerr
			}
		}

		logger.Info(fmt.Sprintf("Received from '%v': '%v'", s.socketName, response))

		if exitMsg != "" && strings.Contains(response, exitMsg) {
			logger.Info(fmt.Sprintf("Received exit msg '%v' from '%v'. Closing client.", exitMsg, s.socketName))
			return response, err
		}
	}
}

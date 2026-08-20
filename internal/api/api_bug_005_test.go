package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"
)

type oneConnListener struct {
	conn net.Conn
	used bool
}

func (l *oneConnListener) Accept() (net.Conn, error) {
	if l.used {
		return nil, errors.New("closed")
	}
	l.used = true
	return l.conn, nil
}
func (l *oneConnListener) Close() error   { l.used = true; return nil }
func (l *oneConnListener) Addr() net.Addr { return pipeAddr("test") }

type pipeAddr string

func (p pipeAddr) Network() string { return "pipe" }
func (p pipeAddr) String() string  { return string(p) }

type contextHandler struct{}

func (contextHandler) Handle(ctx context.Context, _ RPCRequest) (RPCResponse, error) {
	select {
	case <-ctx.Done():
		return RPCResponse{Code: 499}, nil
	default:
		return RPCResponse{Code: 200}, nil
	}
}

func TestLineRPCServeContextCancellation(t *testing.T) {
	server, client := net.Pipe()
	ln := &oneConnListener{conn: server}
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	rpc := LineRPC{Handler: contextHandler{}, Timeout: time.Second}
	done := make(chan error, 1)
	go func() { done <- rpc.ServeContext(parent, ln) }()
	if err := json.NewEncoder(client).Encode(RPCRequest{Method: "ping"}); err != nil {
		t.Fatal(err)
	}
	var resp RPCResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 499 {
		t.Fatalf("parent cancellation lost: %#v", resp)
	}
	_ = client.Close()
	<-done
}

func TestLineRPCHandleContextCancellation(t *testing.T) {
	server, client := net.Pipe()
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	rpc := LineRPC{Handler: contextHandler{}, Timeout: time.Second}
	done := make(chan struct{})
	go func() { rpc.handleContext(parent, server); close(done) }()
	if err := json.NewEncoder(client).Encode(RPCRequest{Method: "ping"}); err != nil { t.Fatal(err) }
	var resp RPCResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil { t.Fatal(err) }
	if resp.Code != 499 { t.Fatalf("direct handler lost cancellation: %#v", resp) }
	_ = client.Close()
	<-done
}

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type RPCRequest struct {
	Method    string          `json:"method"`
	Payload   json.RawMessage `json:"payload"`
	RequestID string          `json:"request_id"`
}
type RPCResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error,omitempty"`
	Payload any    `json:"payload,omitempty"`
}
type RPCHandler interface {
	Handle(context.Context, RPCRequest) (RPCResponse, error)
}
type LineRPC struct {
	Handler RPCHandler
	Timeout time.Duration
}

func (s LineRPC) Serve(l net.Listener) error {
	for {
		c, e := l.Accept()
		if e != nil {
			return e
		}
		go s.handle(c)
	}
}
func (s LineRPC) handle(c net.Conn) {
	defer c.Close()
	dec := json.NewDecoder(c)
	enc := json.NewEncoder(c)
	for {
		var req RPCRequest
		if e := dec.Decode(&req); e != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
		res, e := s.Handler.Handle(ctx, req)
		cancel()
		if e != nil {
			res = RPCResponse{Code: 500, Error: fmt.Sprintf("rpc: %v", e)}
		}
		_ = enc.Encode(res)
	}
}

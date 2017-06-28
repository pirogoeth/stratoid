package stratum

import (
	"bytes"
	"encoding/json"
	"net"

	"github.com/mailgun/holster/errors"
)

type Params map[string]interface{}

type Request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params Params `json:"params"`
}

// DecodeRequest unmarshals a request from the miner client into a
// Request struct.
func DecodeRequest(data []byte) (*Request, error) {
	req := &Request{}

	data = bytes.TrimRight(data, "\x0d\x0a\x00")

	if err := json.Unmarshal(data, req); err != nil {
		return nil, errors.Wrap(err, "while decoding rpc request")
	}

	return req, nil
}

// Encode marshals a request into a Stratum RPC-compatible form for forwarding
// to the pool server.
func (req *Request) Encode() ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "while encoding rpc request")
	}

	return append(data, []byte{0x0a}...), nil
}

func (req *Request) Send(sock net.Conn) error {
	body, err := req.Encode()
	if err != nil {
		return errors.Wrap(err, "while sending request to socket")
	}

	// log.Printf("writing request to socket: %s\n", body)

	_, err = sock.Write(body)
	if err != nil {
		return errors.Wrap(err, "while sending request to socket")
	}

	return nil
}

func (req *Request) WithFields(fields Params) *Request {
	if req.Params == nil {
		req.Params = make(map[string]interface{}, 0)
	}

	for k, v := range fields {
		req.Params[k] = v
	}

	return req
}

type Response struct {
	ID         int                    `json:"id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Params     Params                 `json:"params,omitempty"`
	Result     map[string]interface{} `json:"result"`
	Error      map[string]interface{} `json:"error"`
	RPCVersion string                 `json:"jsonrpc"`
}

// DecodeResponse unmarshals a response from the pool server into a
// Response struct.
func DecodeResponse(data []byte) (*Response, error) {
	resp := &Response{}

	data = bytes.TrimRight(data, "\x0d\x0a\x00")

	if err := json.Unmarshal(data, resp); err != nil {
		return nil, errors.Wrap(err, "while decoding rpc response")
	}

	return resp, nil
}

func (resp *Response) IsCall() bool {
	return resp.Method != ""
}

func (resp *Response) Send(sock net.Conn) error {
	body, err := resp.Encode()
	if err != nil {
		return errors.Wrap(err, "while sending response to socket")
	}

	// log.Printf("writing response to socket: %s\n", body)

	_, err = sock.Write(body)
	if err != nil {
		return errors.Wrap(err, "while sending response to socket")
	}

	return nil
}

// Encode marshals a response into a Stratum RPC-compatible form for forwarding
// to the miner client.
func (resp *Response) Encode() ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.Wrap(err, "while encoding rpc response")
	}

	return append(data, []byte{0x0a}...), nil
}

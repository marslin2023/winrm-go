package winrm

import (
	"crypto/x509"
	"net"
	"net/http"
	"net/url"

	"github.com/Azure/go-ntlmssp"
	"github.com/marslin2023/winrm-go/soap"
)

// ClientNTLM provides a transport via NTLMv2
type ClientNTLM struct {
	clientRequest
}

type ClientNTLMOption func(*clientRequest)

// Transport creates the wrapped NTLM transport
func (c *ClientNTLM) Transport(endpoint *Endpoint) error {
	if err := c.clientRequest.Transport(endpoint); err != nil {
		return err
	}
	c.clientRequest.transport = &ntlmssp.Negotiator{RoundTripper: c.clientRequest.transport}
	return nil
}

// Post make post to the winrm soap service (forwarded to clientRequest implementation)
func (c ClientNTLM) Post(client *Client, request *soap.SoapMessage) (string, error) {
	return c.clientRequest.Post(client, request)
}

// NewClientNTLMWithDial NewClientNTLMWithDial
func NewClientNTLMWithDial(dial func(network, addr string) (net.Conn, error)) *ClientNTLM {
	return &ClientNTLM{
		clientRequest{
			dial: dial,
		},
	}
}

// NewClientNTLMWithProxyFunc NewClientNTLMWithProxyFunc
func NewClientNTLMWithProxyFunc(proxyfunc func(req *http.Request) (*url.URL, error)) *ClientNTLM {
	return &ClientNTLM{
		clientRequest{
			proxyfunc: proxyfunc,
		},
	}
}

func NewClientNTLMWithKeyCheckFunc(keyCheck func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error) *ClientNTLM {
	return &ClientNTLM{
		clientRequest{
			keyCheck: keyCheck,
		},
	}
}


func WithDial(dial func(network, addr string) (net.Conn, error)) ClientNTLMOption {
	return func(cr *clientRequest) {
		cr.dial = dial
	}
}

func WithProxyFunc(proxyfunc func(req *http.Request) (*url.URL, error)) ClientNTLMOption {
	return func(cr *clientRequest) {
		cr.proxyfunc = proxyfunc
	}
}

func WithKeyCheckFunc(keyCheck func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error) ClientNTLMOption {
	return func(cr *clientRequest) {
		cr.keyCheck = keyCheck
	}
}

func NewClientNTLM(opts ...ClientNTLMOption) *ClientNTLM {
	cr := clientRequest{}
	for _, opt := range opts {
		opt(&cr)
	}
	return &ClientNTLM{cr}
}

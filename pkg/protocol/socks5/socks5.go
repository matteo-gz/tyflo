package socks5

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

// ref doc: https://datatracker.ietf.org/doc/html/rfc1928

const (
	Version5                       = 5
	MethodNoAuthenticationRequired = 0x0
	MethodGSSAPI                   = 0x1
	MethodUsernamePassword         = 0x2
	CmdCONNECT                     = 0x01
	CmdBIND                        = 0x02
	CmdUDPAssociate                = 0x03
	RSVDefault                     = 0x00
	ATYPIPV4Address                = 0x01
	ATYPDomainName                 = 0x03
	ATYPIPV6Address                = 0x04
	RepSucceeded                   = 0x0
)

var (
	ErrVersionNotSupport = errors.New("version not support")
	ErrVersionNotV5      = errors.New("version is not 5")
	ErrMethodLen         = errors.New("method len is 0")
	ErrCmdNotSupport     = errors.New("cmd not support")
	ErrRsvInvalid        = errors.New("request rsv invalid")
	ErrATYPInvalid       = errors.New("request ATYP invalid")
	ErrMethodNotSupport  = errors.New("method not support")
)

type ClientNegotiateReq struct {
	Version  byte
	NMethods byte
	Methods  []byte
}

func NewClientNegotiateReq() *ClientNegotiateReq {
	return &ClientNegotiateReq{}
}
func (req *ClientNegotiateReq) Get(r io.Reader) (err error) {
	buf := make([]byte, 2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	req.Version = buf[0]
	if req.Version != Version5 {
		err = ErrVersionNotV5
		return
	}
	req.NMethods = buf[1]
	if req.NMethods <= 0 {
		err = ErrMethodLen
		return
	}
	buf = make([]byte, req.NMethods)
	_, err = io.ReadAtLeast(r, buf, int(req.NMethods))
	if err != nil {
		return
	}
	req.Methods = buf
	return
}
func (req *ClientNegotiateReq) GetNoAuthenticationRequired() []byte {
	req.NMethods = 1
	req.Methods = []byte{MethodNoAuthenticationRequired}
	req.Version = Version5
	data := []byte{req.Version, req.NMethods}
	for _, v := range req.Methods {
		data = append(data, v)
	}
	return data
}

type ServerNegotiateReply struct {
	Version byte
	Method  byte
}

func NewServerNegotiateReply() *ServerNegotiateReply {
	return &ServerNegotiateReply{}
}
func (s *ServerNegotiateReply) Get() []byte {
	return []byte{s.Version, s.Method}
}
func (s *ServerNegotiateReply) SetNotPassword() {
	s.Version = Version5
	s.Method = MethodNoAuthenticationRequired
}
func (s *ServerNegotiateReply) Decode(ctx context.Context, r io.Reader) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	s.Version = buf[0]
	if s.Version != Version5 {
		return ErrVersionNotV5
	}
	s.Method = buf[1]
	return nil
}

type ClientRequest struct {
	VER     byte
	CMD     byte
	RSV     byte
	ATYP    byte
	DSTAddr []byte
	DSTPort uint16
	host    string
}

func NewClientRequest() *ClientRequest {
	return &ClientRequest{}
}
func (c *ClientRequest) String() string {
	return fmt.Sprintf(
		"ver %v cmd %v rsv %v atyp %v dst addr %v dst port %v host %v",
		c.VER,
		c.CMD,
		c.RSV,
		c.ATYP,
		c.DSTAddr,
		c.DSTPort,
		c.host,
	)
}
func (c *ClientRequest) Get(ctx context.Context, r io.Reader) (err error) {
	buf := make([]byte, 4)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	c.VER = buf[0]
	if c.VER != Version5 {
		return ErrVersionNotV5
	}
	c.CMD = buf[1]
	switch c.CMD {
	case CmdCONNECT:
	case CmdBIND:
	case CmdUDPAssociate:
	default:
		return ErrCmdNotSupport
	}
	c.RSV = buf[2]
	if c.RSV != RSVDefault {
		return ErrRsvInvalid
	}
	c.ATYP = buf[3]
	var addrLen int
	switch c.ATYP {
	case ATYPIPV4Address:
		addrLen = net.IPv4len
	case ATYPDomainName:
		addrLen = int(buf[4])
	case ATYPIPV6Address:
		addrLen = net.IPv6len
	default:
		return ErrATYPInvalid
	}
	buf = make([]byte, addrLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	if c.ATYP == ATYPDomainName {
		c.host = string(buf)
	} else {
		c.host = net.IP(buf).String()
	}
	c.DSTAddr = buf
	buf = make([]byte, 2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	c.DSTPort = binary.BigEndian.Uint16(buf)
	return nil
}
func (c *ClientRequest) GetAddress() string {
	return net.JoinHostPort(c.host, strconv.Itoa(int(c.DSTPort)))
}

type ServerReply struct {
	VER     byte
	REP     byte
	RSV     byte
	ATYP    byte
	BNDAddr []byte
	BNDPort uint16
}

func NewServerReply() *ServerReply {
	return &ServerReply{}
}
func (s *ServerReply) Get() []byte {
	d := []byte{
		s.VER,
		s.REP,
		s.RSV,
		s.ATYP,
	}
	for _, v := range s.BNDAddr {
		d = append(d, v)
	}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, s.BNDPort)
	d = append(d, b...)
	return d
}

// SetConnectDirectReply not bind address
func (s *ServerReply) SetConnectDirectReply() {
	s.VER = Version5
	s.REP = RepSucceeded
	s.RSV = RSVDefault
	s.ATYP = ATYPIPV4Address
	for i := 0; i < net.IPv4len; i++ {
		s.BNDAddr = append(s.BNDAddr, 0)
	}
	s.BNDPort = 0
}

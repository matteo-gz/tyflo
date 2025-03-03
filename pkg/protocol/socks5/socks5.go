package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
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
	ErrBadVersion        = errors.New("bad version")
	ErrMethodLen         = errors.New("method len is 0")
	ErrCmdNotSupport     = errors.New("cmd not support")
	ErrRsvInvalid        = errors.New("request rsv invalid")
	ErrATYPInvalid       = errors.New("request ATYP invalid")
	ErrMethodNotSupport  = errors.New("method not support")
	ErrHostInvalid       = errors.New("host invalid")
	ErrUserPasswordLen   = errors.New("user or password len over 255")
)

type Message interface {
	Decode(r io.Reader) (err error)
	Bytes() []byte
}
type ClientNegotiateReq struct {
	Version  byte
	NMethods byte
	Methods  []byte
}

func NewClientNegotiateReq() *ClientNegotiateReq {
	return &ClientNegotiateReq{}
}
func (req *ClientNegotiateReq) Decode(r io.Reader) (err error) {
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
func (req *ClientNegotiateReq) SetNoAuthenticationRequired() {
	req.Methods = []byte{MethodNoAuthenticationRequired}
	req.NMethods = byte(len(req.Methods))
	req.Version = Version5
}
func (req *ClientNegotiateReq) SetUsernamePassword() {
	req.Methods = []byte{MethodNoAuthenticationRequired, MethodUsernamePassword}
	req.NMethods = byte(len(req.Methods))
	req.Version = Version5
}
func (req *ClientNegotiateReq) Bytes() []byte {
	data := []byte{req.Version, req.NMethods}
	data = append(data, req.Methods...)
	return data
}

type ServerNegotiateReply struct {
	Version byte
	Method  byte
}

func NewServerNegotiateReply() *ServerNegotiateReply {
	return &ServerNegotiateReply{}
}
func (s *ServerNegotiateReply) Bytes() []byte {
	return []byte{s.Version, s.Method}
}
func (s *ServerNegotiateReply) SetNotPassword() {
	s.Version = Version5
	s.Method = MethodNoAuthenticationRequired
}
func (s *ServerNegotiateReply) SetUsernamePassword() {
	s.Version = Version5
	s.Method = MethodUsernamePassword
}
func (s *ServerNegotiateReply) Decode(r io.Reader) error {
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
func (c *ClientRequest) Decode(r io.Reader) (err error) {
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
		buf = make([]byte, 1)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return
		}
		addrLen = int(buf[0])
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
func (c *ClientRequest) SetCmdConnect(address string) (err error) {
	c.VER = Version5
	c.CMD = CmdCONNECT
	c.RSV = RSVDefault
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return
	}
	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		return
	}
	c.DSTPort = uint16(portNum)
	if ip := net.ParseIP(host); ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			c.ATYP = ATYPIPV4Address
			c.DSTAddr = ipv4
		} else {
			c.ATYP = ATYPIPV6Address
			c.DSTAddr = ip.To16()
		}
	} else {
		c.ATYP = ATYPDomainName
		hl := len(host)
		if hl > math.MaxUint8 {
			err = ErrHostInvalid
			return
		}
		c.DSTAddr = append(c.DSTAddr, uint8(hl))
		c.DSTAddr = append(c.DSTAddr, []byte(host)...)
	}
	return nil
}
func (c *ClientRequest) Bytes() []byte {
	d := []byte{
		c.VER,
		c.CMD,
		c.RSV,
		c.ATYP,
	}
	d = append(d, c.DSTAddr...)
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, c.DSTPort)
	d = append(d, b...)
	return d
}

type ServerReply struct {
	VER     byte
	REP     byte
	RSV     byte
	ATYP    byte
	BNDAddr []byte
	BNDPort uint16
	host    string
}

func NewServerReply() *ServerReply {
	return &ServerReply{}
}
func (s *ServerReply) Bytes() []byte {
	d := []byte{
		s.VER,
		s.REP,
		s.RSV,
		s.ATYP,
	}
	d = append(d, s.BNDAddr...)
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
func (s *ServerReply) Decode(r io.Reader) (err error) {
	buf := make([]byte, 4)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	s.VER = buf[0]
	if s.VER != Version5 {
		return ErrVersionNotV5
	}
	s.REP = buf[1]
	s.RSV = buf[2]
	if s.RSV != RSVDefault {
		return ErrRsvInvalid
	}
	s.ATYP = buf[3]
	var addrLen int
	switch s.ATYP {
	case ATYPIPV4Address:
		addrLen = net.IPv4len
	case ATYPDomainName:
		buf = make([]byte, 1)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return
		}
		addrLen = int(buf[0])
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
	if s.ATYP == ATYPDomainName {
		s.host = string(buf)
	} else {
		s.host = net.IP(buf).String()
	}
	s.BNDAddr = buf
	buf = make([]byte, 2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	s.BNDPort = binary.BigEndian.Uint16(buf)
	return nil
}

type UsernamePasswordReq struct {
	VER    byte
	ULEN   byte
	UNAME  string
	PLEN   byte
	PASSWD string
}

func (req *UsernamePasswordReq) Bytes() []byte {
	data := []byte{req.VER, req.ULEN}
	data = append(data, req.UNAME...)
	data = append(data, req.PLEN)
	data = append(data, req.PASSWD...)
	return data
}
func (req *UsernamePasswordReq) Decode(r io.Reader) (err error) {
	// 读取版本和用户名长度
	buf := make([]byte, 2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	req.VER = buf[0]
	if req.VER != UserPasswordVersion {
		return ErrBadVersion
	}
	req.ULEN = buf[1]
	// 读取用户名
	buf = make([]byte, req.ULEN)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	req.UNAME = string(buf)
	// 读取密码长度
	buf = make([]byte, 1)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	req.PLEN = buf[0]
	// 读取密码
	buf = make([]byte, req.PLEN)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	req.PASSWD = string(buf)
	return nil
}
func (req *UsernamePasswordReq) SetUsernamePassword(user, password string) error {
	var l1, l2 int
	l1 = len(user)
	l2 = len(password)
	if l1 > 255 {
		return ErrUserPasswordLen
	}
	if l2 > 255 {
		return ErrUserPasswordLen
	}
	req.VER = UserPasswordVersion
	req.ULEN = byte(l1)
	req.UNAME = user
	req.PLEN = byte(l2)
	req.PASSWD = password
	return nil
}
func NewUsernamePasswordReq() *UsernamePasswordReq {
	return &UsernamePasswordReq{}
}

type UsernamePasswordReply struct {
	VER    byte
	STATUS byte
}

func (reply *UsernamePasswordReply) Bytes() []byte {
	return []byte{reply.VER, reply.STATUS}
}
func (reply *UsernamePasswordReply) SetSuccess() {
	reply.VER = UserPasswordVersion
	reply.STATUS = UserPasswordOk
}
func (reply *UsernamePasswordReply) SetFailure() {
	reply.VER = UserPasswordVersion
	reply.STATUS = UserPasswordFailure
}
func (reply *UsernamePasswordReply) Decode(r io.Reader) (err error) {
	buf := make([]byte, 2)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return fmt.Errorf("UsernamePasswordReply.Decode len:%d %w", n, err)
	}
	reply.VER = buf[0]
	reply.STATUS = buf[1]
	if reply.VER != UserPasswordVersion {
		return ErrBadVersion
	}
	if reply.STATUS != UserPasswordOk {
		return ErrReplyFail
	}
	return nil
}
func NewUsernamePasswordReply() *UsernamePasswordReply {
	return &UsernamePasswordReply{}
}

const (
	UserPasswordVersion = 1
	UserPasswordOk      = 0
	UserPasswordFailure = 1
)

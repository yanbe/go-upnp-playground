package ssdp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

const (
	DefaultMaxMessageBytes = 2 << 10
)

var (
	crlf = []byte("\r\n")

	bufBytesPool      sync.Pool
	bufioReaderPool   sync.Pool
	bufioWriter2kPool sync.Pool
	bufioWriter4kPool sync.Pool
)

// Handler is the interface by which received HTTPU messages are passed to
// handling code.
type Handler interface {
	// ServeMessage is called for each HTTPU message received. peerAddr contains
	// the address that the message was received from.
	ServeMessage(w http.ResponseWriter, r *http.Request)
}

// A Server defines parameters for running an HTTPU server.
type SSDPDiscoveryResponder struct {
	serviceAddr     net.TCPAddr    // TCP address to listen on
	Multicast       bool           // Should listen for multicast?
	Interface       *net.Interface // Network interface to listen on for multicast, nil for default multicast interface
	Handler         Handler        // handler to invoke
	MaxMessageBytes int            // maximum number of bytes to read from a packet, DefaultMaxMessageBytes if 0
	deviceUUID      uuid.UUID
}

// ListenAndServe listens on the UDP network address srv.Addr. If srv.Multicast
// is true, then a multicast UDP listener will be used on srv.Interface (or
// default interface if nil).
func (s *SSDPDiscoveryResponder) ListenAndServe() error {
	var err error

	var listenAddr *net.UDPAddr
	if listenAddr, err = net.ResolveUDPAddr("udp", ssdpUDP4Addr); err != nil {
		log.Fatal(err)
	}

	var conn net.PacketConn
	if s.Multicast {
		if conn, err = net.ListenMulticastUDP("udp", s.Interface, listenAddr); err != nil {
			return err
		}
	} else {
		if conn, err = net.ListenUDP("udp", listenAddr); err != nil {
			return err
		}
	}

	return s.Serve(conn)
}

func (s *SSDPDiscoveryResponder) newBytesBuf() []byte {
	if v := bufBytesPool.Get(); v != nil {
		buf := v.(*[]byte)
		return *buf
	}
	var maxMessageBytes int
	switch s.MaxMessageBytes {
	case 0:
		maxMessageBytes = DefaultMaxMessageBytes
	default:
		maxMessageBytes = s.MaxMessageBytes
	}
	return make([]byte, maxMessageBytes)
}

func putBytesBuf(buf []byte) {
	bufBytesPool.Put(&buf)
}

func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	// Note: if this reader size is ever changed, update
	// TestHandlerBodyClose's assumptions.
	return bufio.NewReader(r)
}

func putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bufioReaderPool.Put(br)
}

func bufioWriterPool(size int) *sync.Pool {
	switch size {
	case 2 << 10:
		return &bufioWriter2kPool
	case 4 << 10:
		return &bufioWriter4kPool
	}
	return nil
}

func newBufioWriterSize(w io.Writer, size int) *bufio.Writer {
	pool := bufioWriterPool(size)
	if pool != nil {
		if v := pool.Get(); v != nil {
			bw := v.(*bufio.Writer)
			bw.Reset(w)
			return bw
		}
	}
	return bufio.NewWriterSize(w, size)
}

func putBufioWriter(bw *bufio.Writer) {
	bw.Reset(nil)
	if pool := bufioWriterPool(bw.Available()); pool != nil {
		pool.Put(bw)
	}
}

type UDPResponseWriter struct {
	conn         net.PacketConn
	addr         net.Addr
	req          *http.Request
	res          *response
	header       *http.Header
	calledHeader bool
	wroteHeader  bool
	status       int
	bufw         *bufio.Writer
	statusBuf    [3]byte
}

type response struct {
	rw *UDPResponseWriter
}

func (r *response) Write(data []byte) (int, error) {
	log.Print(string(data))
	return r.rw.conn.WriteTo(data, r.rw.addr)
}

func (w *UDPResponseWriter) Header() http.Header {
	if !w.calledHeader {
		w.header = &http.Header{}
	}
	w.calledHeader = true
	return *w.header
}

// TODO: http.Server compatible implementation
func (w *UDPResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.bufw.Write(body)
}

func checkWriteHeaderCode(code int) {
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

func writeStatusLine(bw *bufio.Writer, is11 bool, code int, scratch []byte) {
	if is11 {
		bw.WriteString("HTTP/1.1 ")
	} else {
		bw.WriteString("HTTP/1.0 ")
	}
	text := http.StatusText(code)
	if text != "" {
		bw.Write(strconv.AppendInt(scratch[:0], int64(code), 10))
		bw.WriteByte(' ')
		bw.WriteString(text)
		bw.WriteString("\r\n")
	} else {
		// don't worry about performance
		fmt.Fprintf(bw, "%03d status code %d\r\n", code, code)
	}
}
func (w *UDPResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		log.Default().Fatal("httpu: superfluous response.WriteHeader")
	}
	if !w.calledHeader {
		w.Header() // instantiate
	}
	checkWriteHeaderCode(code)
	w.wroteHeader = true
	w.status = code
	writeStatusLine(w.bufw, w.req.ProtoAtLeast(1, 1), code, w.statusBuf[:])
	w.header.WriteSubset(w.bufw, nil)
	w.bufw.Write(crlf)
}

func (w *UDPResponseWriter) finishRequest() {
	defer putBufioWriter(w.bufw)
	if !w.wroteHeader {
		if !w.calledHeader {
			return // unsupported or invalid request
		}
		w.WriteHeader(http.StatusOK)
	}
	w.bufw.Flush()
}

// Serve messages received on the given packet listener to the srv.Handler.
func (s *SSDPDiscoveryResponder) Serve(l net.PacketConn) error {
	for {
		buf := s.newBytesBuf()
		n, addr, err := l.ReadFrom(buf)
		if err != nil {
			return err
		}

		go func(buf []byte, n int, addr net.Addr) {
			r := io.LimitReader(bytes.NewReader(buf), int64(n))
			br := newBufioReader(r)
			req, err := http.ReadRequest(br)
			putBytesBuf(buf)
			putBufioReader(br)
			if err != nil {
				log.Printf("httpu: Failed to parse request: %v", err)
				return
			}
			req.RemoteAddr = addr.String()
			rw := &UDPResponseWriter{
				conn: l,
				addr: addr,
				req:  req,
			}
			rw.res = &response{
				rw,
			}
			rw.bufw = newBufioWriterSize(rw.res, 2<<10)
			s.ServeMessage(rw, req)
			rw.finishRequest()
		}(buf, n, addr)
	}
}

type UDPRoundTripper struct {
}

func (t *UDPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	req.Write(&buf)
	destAddr, err := net.ResolveUDPAddr("udp", req.Host)
	if err != nil {
		return nil, err
	}
	conn.WriteTo(buf.Bytes(), destAddr)
	return &http.Response{}, nil
}

package connfu_test

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/mmatczuk/connfu"
)

type connWrapper struct {
	net.Conn
	rx atomic.Uint64
}

func (cw *connWrapper) Read(b []byte) (n int, err error) {
	n, err = cw.Conn.Read(b)
	cw.rx.Add(uint64(n))
	return
}

func (cw *connWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = cw.Conn.(io.ReaderFrom).ReadFrom(r)
	cw.rx.Add(uint64(n))
	return
}

func (cw *connWrapper) Rx() uint64 {
	return cw.rx.Load()
}

func Example() {
	w := &connWrapper{
		// Do something with the connection
	}

	{
		conn := connfu.Combine(w, new(net.TCPConn))
		if _, ok := conn.(io.ReaderFrom); ok {
			fmt.Println("Combined with net.TCPConn implements io.ReaderFrom")
		} else {
			fmt.Println("Combined with net.TCPConn does not implement io.ReaderFrom")
		}
		if _, ok := conn.(io.WriterTo); ok {
			fmt.Println("Combined with net.TCPConn implements io.WriterTo")
		} else {
			fmt.Println("Combined with net.TCPConn does not implement io.WriterTo")
		}
	}

	{
		conn := connfu.Combine(w, nil)
		if _, ok := conn.(io.ReaderFrom); ok {
			fmt.Println("Combined with nil implements io.ReaderFrom")
		} else {
			fmt.Println("Combined with nil does not implement io.ReaderFrom")
		}
		if _, ok := conn.(io.WriterTo); ok {
			fmt.Println("Combined with nil implements io.WriterTo")
		} else {
			fmt.Println("Combined with nil does not implement io.WriterTo")
		}
	}

	// Output:
	// Combined with net.TCPConn implements io.ReaderFrom
	// Combined with net.TCPConn implements io.WriterTo
	// Combined with nil does not implement io.ReaderFrom
	// Combined with nil does not implement io.WriterTo
}

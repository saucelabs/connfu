package connfu

import (
	"errors"
	"io"
	"net"
	"testing"
)

type testConn struct {
	net.Conn
}

var testAddr = &net.TCPAddr{}

func (tc testConn) RemoteAddr() net.Addr {
	return testAddr
}

func TestCombinePreservesInnerInterface(t *testing.T) {
	testCombinePreservesInnerInterface(t, testConn{})
}

func TestCombinePreservesInnerInterfaceRegardlessOfOuterInterface(t *testing.T) {
	testCombinePreservesInnerInterface(t, struct {
		net.Conn
		io.ReaderFrom
		io.WriterTo
		_closeWriter
	}{testConn{}, nopReaderFrom, nopWriteTo, nopCloseWriter})
}

func testCombinePreservesInnerInterface(t *testing.T, tc net.Conn) {
	t.Helper()
	tests := []struct {
		name         string
		outer, inner net.Conn
		flags        uint8
	}{
		{
			name:  "net.Conn",
			outer: tc,
			inner: testConn{},
			flags: 0,
		},
		{
			name:  "ReaderFrom",
			outer: tc,
			inner: struct {
				net.Conn
				io.ReaderFrom
			}{tc, nopReaderFrom},
			flags: readerFrom,
		},
		{
			name:  "WriterTo",
			outer: tc,
			inner: struct {
				net.Conn
				io.WriterTo
			}{tc, nopWriteTo},
			flags: writerTo,
		},
		{
			name:  "CloseWriter",
			outer: tc,
			inner: struct {
				net.Conn
				_closeWriter
			}{tc, nopCloseWriter},
			flags: closeWriter,
		},
		{
			name:  "ReaderFrom+WriterTo",
			outer: tc,
			inner: struct {
				net.Conn
				io.ReaderFrom
				io.WriterTo
			}{tc, nopReaderFrom, nopWriteTo},
			flags: readerFrom | writerTo,
		},
		{
			name:  "ReaderFrom+WriterTo+CloseWriter",
			outer: tc,
			inner: struct {
				net.Conn
				io.ReaderFrom
				io.WriterTo
				_closeWriter
			}{tc, nopReaderFrom, nopWriteTo, nopCloseWriter},
			flags: readerFrom | writerTo | closeWriter,
		},
	}

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			conn := Combine(tc.outer, tc.inner)

			if got := conn.RemoteAddr(); got != testAddr {
				t.Fatalf("RemoteAddr() = %v, want %v", got, testAddr)
			}

			if got := flags(conn); got != tc.flags {
				t.Fatalf("flags(conn) = %d, want %d", got, tc.flags)
			}
			if _, ok := conn.(io.ReaderFrom); (tc.flags&readerFrom != 0) != ok {
				t.Fatal("type assertion failed for io.ReaderFrom")
			}
			if _, ok := conn.(io.WriterTo); (tc.flags&writerTo != 0) != ok {
				t.Fatal("type assertion failed for io.WriterTo")
			}
			if _, ok := conn.(_closeWriter); (tc.flags&closeWriter != 0) != ok {
				t.Fatal("type assertion failed for _closeWriter")
			}
		})
	}
}

func TestCombineOverloading(t *testing.T) {
	tc := testConn{}
	testErr := errors.New("test error")

	tests := []struct {
		name         string
		outer, inner net.Conn
	}{
		{
			name: "ReaderFrom",
			outer: struct {
				net.Conn
				io.ReaderFrom
			}{tc, readerFromFunc(func(r io.Reader) (int64, error) {
				return 42, testErr
			})},
			inner: struct {
				net.Conn
				io.ReaderFrom
			}{tc, nopReaderFrom},
		},
		{
			name: "WriterTo",
			outer: struct {
				net.Conn
				io.WriterTo
			}{tc, writeToFunc(func(w io.Writer) (int64, error) {
				return 42, testErr
			})},
			inner: struct {
				net.Conn
				io.WriterTo
			}{tc, nopWriteTo},
		},
		{
			name: "CloseWriter",
			outer: struct {
				net.Conn
				_closeWriter
			}{tc, closeWriterFunc(func() error {
				return testErr
			})},
			inner: struct {
				net.Conn
				_closeWriter
			}{tc, nopCloseWriter},
		},
	}

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			conn := Combine(tc.outer, tc.inner)

			if got := conn.RemoteAddr(); got != testAddr {
				t.Fatalf("RemoteAddr() = %v, want %v", got, testAddr)
			}

			flags := flags(conn)
			if v, ok := conn.(io.ReaderFrom); (flags&readerFrom != 0) != ok {
				t.Fatal("type assertion failed for io.ReaderFrom")
			} else if ok {
				if _, err := v.ReadFrom(nil); !errors.Is(err, testErr) {
					t.Fatalf("ReadFrom() = %v, want %v", err, testErr)
				}
			}
			if v, ok := conn.(io.WriterTo); (flags&writerTo != 0) != ok {
				t.Fatal("type assertion failed for io.WriterTo")
			} else if ok {
				if _, err := v.WriteTo(nil); !errors.Is(err, testErr) {
					t.Fatalf("WriteTo() = %v, want %v", err, testErr)
				}
			}
			if v, ok := conn.(_closeWriter); (flags&closeWriter != 0) != ok {
				t.Fatal("type assertion failed for _closeWriter")
			} else if ok {
				if err := v.CloseWrite(); !errors.Is(err, testErr) {
					t.Fatalf("CloseWrite() = %v, want %v", err, testErr)
				}
			}
		})
	}
}

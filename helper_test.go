package connfu

import (
	"io"
)

type readerFromFunc func(r io.Reader) (int64, error)

func (r readerFromFunc) ReadFrom(rd io.Reader) (int64, error) {
	return r(rd)
}

var nopReaderFrom io.ReaderFrom = readerFromFunc(func(io.Reader) (int64, error) {
	return 0, nil
})

type writeToFunc func(w io.Writer) (int64, error)

func (wt writeToFunc) WriteTo(w io.Writer) (int64, error) {
	return wt(w)
}

var nopWriteTo io.WriterTo = writeToFunc(func(io.Writer) (int64, error) {
	return 0, nil
})

type closeWriterFunc func() error

func (cw closeWriterFunc) CloseWrite() error {
	return cw()
}

var _ _closeWriter = closeWriterFunc(nil)

var nopCloseWriter _closeWriter = closeWriterFunc(func() error {
	return nil
})

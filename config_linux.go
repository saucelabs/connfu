//go:build linux

package connfu

func DefaultConfig() Config {
	return Config{
		UseReaderFrom: true,
		UseWriterTo:   false, // spliceTo optimization requires that the provided writer is a UnixConn ptr - disable in general by default
	}
}

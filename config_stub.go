//go:build !linux

package connfu

func DefaultConfig() Config {
	return Config{}
}

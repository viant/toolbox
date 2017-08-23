package bridge

import (
	"io"
	"net/http/httputil"
)

type bytesBufferPool struct {
	channel    chan []byte
	bufferSize int
}

func (p *bytesBufferPool) Get() (result []byte) {
	select {
	case result = <-p.channel:
	default:
		result = make([]byte, p.bufferSize)
	}
	return result
}

func (p *bytesBufferPool) Put(b []byte) {
	select {
	case p.channel <- b:
	default: //If the pool is full, discard the buffer.
	}
}

//NewBytesBufferPool returns new httputil.BufferPool pool.
func NewBytesBufferPool(poolSize, bufferSize int) httputil.BufferPool {
	return &bytesBufferPool{
		channel:    make(chan []byte, poolSize),
		bufferSize: bufferSize,
	}
}

//CopyBuffer copies bytes from passed in source to destination with provided pool
func CopyWithBufferPool(source io.Reader, destination io.Writer, pool httputil.BufferPool) (int64, error) {
	buf := pool.Get()
	defer pool.Put(buf)
	var written int64
	for {
		bytesRead, readError := source.Read(buf)
		if bytesRead > 0 {
			bytesWritten, writeError := destination.Write(buf[:bytesRead])
			if bytesWritten > 0 {
				written += int64(bytesWritten)
			}
			if writeError != nil {
				return written, writeError
			}
			if bytesRead != bytesWritten {
				return written, io.ErrShortWrite
			}
		}
		if readError != nil {
			return written, readError
		}
	}
}

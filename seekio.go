// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package seekio provides Seeker functionality for Readers and Writers.
package tiff

import (
	"errors"
	"fmt"
	"io"
)

var (
	_ seekReadCloser  = (*seekioReader)(nil)
	_ seekWriteCloser = (*seekioWriter)(nil)

	ErrInvalidWhence  = errors.New("seekio: invalid whence")
	ErrNegativeOffset = errors.New("seekio: negative offset")
	ErrIntOverflow    = errors.New("seekio: int64 to int conversion overflow")
)

// seekReadCloser combines io.Seeker, io.Reader, and io.Closer interfaces.
type seekReadCloser interface {
	io.Seeker
	io.Reader
	io.Closer
}

// seekWriteCloser combines io.Seeker, io.Writer, and io.Closer interfaces.
type seekWriteCloser interface {
	io.Seeker
	io.Writer
	io.Closer
}

// seekioReader implements seekReadCloser for any io.Reader.
// It buffers all data in memory if the underlying reader is not already a seeker.
type seekioReader struct {
	r   io.Reader     // original reader
	rs  io.ReadSeeker // used if original reader was already a ReadSeeker
	buf []byte        // buffer if original reader was not a ReadSeeker
	off int           // current offset in buffer
	err error         // stored error
}

// NewSeekReader creates a new seekable reader from an existing reader.
// If maxBufferSize > 0, it limits the maximum amount of data read into memory.
func NewSeekReader(r io.Reader, maxBufferSize int) io.ReadSeekCloser {
	// If r already implements ReadSeeker, use it directly
	if rs, ok := r.(io.ReadSeeker); ok {
		return &seekioReader{rs: rs}
	}

	// Otherwise, read all data into memory
	data, err := io.ReadAll(r)
	if err != nil {
		return &seekioReader{err: err}
	}

	// If maxBufferSize is specified and exceeded, return an error
	if maxBufferSize > 0 && len(data) > maxBufferSize {
		return &seekioReader{err: fmt.Errorf("seekio: buffer size %d exceeds maximum %d", len(data), maxBufferSize)}
	}

	return &seekioReader{r: r, buf: data}
}

func (p *seekioReader) Read(data []byte) (n int, err error) {
	if p.err != nil {
		return 0, p.err
	}

	if p.rs != nil {
		return p.rs.Read(data)
	}

	if p.off >= len(p.buf) {
		return 0, io.EOF
	}

	n = copy(data, p.buf[p.off:])
	p.off += n
	return n, nil
}

func (p *seekioReader) Seek(offset int64, whence int) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}

	if p.rs != nil {
		return p.rs.Seek(offset, whence)
	}

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(p.off) + offset
	case io.SeekEnd:
		newOffset = int64(len(p.buf)) + offset
	default:
		return int64(p.off), ErrInvalidWhence
	}

	if newOffset < 0 {
		return int64(p.off), ErrNegativeOffset
	}

	// Check for int overflow on 32-bit platforms
	if int64(int(newOffset)) != newOffset {
		return int64(p.off), ErrIntOverflow
	}

	p.off = int(newOffset)

	// If past EOF, next Read will return EOF
	if p.off > len(p.buf) {
		p.off = len(p.buf)
	}

	return newOffset, nil
}

func (p *seekioReader) Close() error {
	if closer, ok := p.r.(io.Closer); ok && p.rs == nil {
		return closer.Close()
	}

	if closer, ok := p.rs.(io.Closer); ok {
		return closer.Close()
	}

	// Clear buffer to help GC
	p.buf = nil
	return p.err
}

// seekioWriter implements seekWriteCloser for any io.Writer.
type seekioWriter struct {
	w   io.Writer      // original writer
	ws  io.WriteSeeker // used if original writer was already a WriteSeeker
	buf []byte         // buffer if original writer was not a WriteSeeker
	off int            // current offset in buffer
	err error          // stored error
}

// NewSeekWriter creates a new seekable writer from an existing writer.
// If maxBufferSize > 0, it limits the maximum buffer size in memory.
func NewSeekWriter(w io.Writer, maxBufferSize int) (seekWriteCloser, error) {
	if ws, ok := w.(io.WriteSeeker); ok {
		return &seekioWriter{ws: ws}, nil
	}

	initialSize := 4096 // Start with a reasonable buffer size
	if maxBufferSize > 0 && initialSize > maxBufferSize {
		initialSize = maxBufferSize
	}

	return &seekioWriter{
		w:   w,
		buf: make([]byte, 0, initialSize),
	}, nil
}

func (p *seekioWriter) Write(data []byte) (n int, err error) {
	if p.err != nil {
		return 0, p.err
	}

	if p.ws != nil {
		return p.ws.Write(data)
	}

	if err = p.grow(p.off + len(data)); err != nil {
		return 0, err
	}

	n = copy(p.buf[p.off:], data)
	p.off += n
	return n, nil
}

func (p *seekioWriter) grow(n int) error {
	if n <= cap(p.buf) {
		// Enough capacity, just extend the slice
		if n > len(p.buf) {
			p.buf = p.buf[:n]
		}
		return nil
	}

	// Need to allocate more capacity
	newCap := 2 * cap(p.buf)
	if newCap < n {
		newCap = n
	}

	newBuf := make([]byte, len(p.buf), newCap)
	copy(newBuf, p.buf)
	p.buf = newBuf

	if n > len(p.buf) {
		p.buf = p.buf[:n]
	}

	return nil
}

func (p *seekioWriter) Seek(offset int64, whence int) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}

	if p.ws != nil {
		return p.ws.Seek(offset, whence)
	}

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(p.off) + offset
	case io.SeekEnd:
		newOffset = int64(len(p.buf)) + offset
	default:
		return int64(p.off), ErrInvalidWhence
	}

	if newOffset < 0 {
		return int64(p.off), ErrNegativeOffset
	}

	// Check for int overflow on 32-bit platforms
	if int64(int(newOffset)) != newOffset {
		return int64(p.off), ErrIntOverflow
	}

	if err := p.grow(int(newOffset)); err != nil {
		return int64(p.off), err
	}

	p.off = int(newOffset)
	return newOffset, nil
}

func (p *seekioWriter) Close() error {
	if p.err != nil {
		return p.err
	}

	if p.ws != nil {
		if closer, ok := p.ws.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}

	// Write buffered data to the underlying writer
	if len(p.buf) > 0 {
		if _, err := p.w.Write(p.buf[:p.off]); err != nil {
			return err
		}
	}

	// Release buffer
	p.buf = nil

	// Close the underlying writer if it's a closer
	if closer, ok := p.w.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

// openSeekioReader creates a new seekable reader from an existing reader.
// It's a wrapper around NewSeekReader that returns the concrete type.
// If maxBufferSize is negative, there is no limit on the buffer size.
func openSeekioReader(r io.Reader, maxBufferSize int) *seekioReader {
	// If already a seekioReader, just return it
	if sr, ok := r.(*seekioReader); ok {
		return sr
	}

	// If maxBufferSize is negative, there's no limit
	if maxBufferSize < 0 {
		maxBufferSize = 0 // 0 means no limit in NewSeekReader
	}

	// Create a new seekable reader
	readSeeker := NewSeekReader(r, maxBufferSize)

	// Type assertion to get the concrete type
	if sr, ok := readSeeker.(*seekioReader); ok {
		return sr
	}

	// This should never happen if NewSeekReader is implemented correctly,
	// but we need to handle it for safety
	panic("tiff: NewSeekReader did not return a *seekioReader")
}

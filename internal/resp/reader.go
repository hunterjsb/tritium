package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	r *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) ReadValue() (interface{}, error) {
	// Read type byte
	typ, err := r.r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read type error: %w", err)
	}

	switch typ {
	case SimpleString:
		return r.readSimpleString()
	case Error:
		return nil, r.readError()
	case Integer:
		return r.readInteger()
	case BulkString:
		return r.readBulkString()
	case Array:
		return r.readArray()
	default:
		return nil, ErrInvalidResp
	}
}

// ReadBulk reads a value expecting it to be a bulk string
func (r *Reader) ReadBulk() ([]byte, error) {
	// Verify we're getting a bulk string
	typ, err := r.r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read type error: %w", err)
	}
	if typ != BulkString {
		return nil, fmt.Errorf("expected bulk string reply ($), got %c", typ)
	}

	return r.readBulkString()
}

// ReadStr is a convenience method that returns the bulk string as a string
func (r *Reader) ReadStr() (string, error) {
	data, err := r.ReadBulk()
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", nil // handle null bulk string
	}
	return string(data), nil
}

// IsOK reads a simple string "OK" response from the buffer
func (r *Reader) IsOK() bool {
	typ, byteErr := r.r.ReadByte()
	res, readErr := r.readSimpleString()
	if byteErr != nil || readErr != nil || typ != SimpleString {
		return false
	}
	return res == "OK"
}

// ReadInt reads an integer from the buffer
func (r *Reader) ReadInt() (int64, error) {
	typ, byteErr := r.r.ReadByte()
	res, readErr := r.readInteger()
	if byteErr != nil || readErr != nil || typ != Integer {
		return 0, fmt.Errorf("could not read int (as %x); BE: %s; RE: %s", typ, byteErr, readErr)
	}
	return res, nil
}

// internal parsing functions

func (r *Reader) readLine() ([]byte, error) {
	line, err := r.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errors.New("invalid line ending")
	}
	return line[:len(line)-2], nil
}

func (r *Reader) readInteger() (int64, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(line), 10, 64)
}

func (r *Reader) readError() error {
	line, err := r.readLine()
	if err != nil {
		return err
	}
	return errors.New(string(line))
}

func (r *Reader) readSimpleString() (string, error) {
	line, err := r.readLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (r *Reader) readBulkString() ([]byte, error) {
	// Read length
	length, err := r.readInteger()
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, nil // Null bulk string
	}

	// Read string data
	data := make([]byte, length)
	_, err = io.ReadFull(r.r, data)
	if err != nil {
		return nil, err
	}

	// Read trailing \r\n
	crlf := make([]byte, 2)
	_, err = io.ReadFull(r.r, crlf)
	if err != nil {
		return nil, err
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return nil, errors.New("invalid bulk string termination")
	}

	return data, nil
}

func (r *Reader) readArray() ([]interface{}, error) {
	// Read array length
	length, err := r.readInteger()
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, nil // Null array
	}

	// Read array elements
	array := make([]interface{}, length)
	for i := int64(0); i < length; i++ {
		value, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		array[i] = value
	}

	return array, nil
}

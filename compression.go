package traefik_umami_plugin

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
)

const (
	// Gzip compression algorithm string.
	Gzip string = "gzip"
	// Deflate compression algorithm string.
	Deflate string = "deflate"
	// Identity compression algorithm string.
	Identity string = "identity"
)

// ReaderError for notating that an error occurred while reading compressed data.
type ReaderError struct {
	error

	cause error
}

// Decode data in a bytes.Reader based on supplied encoding.
func Decode(byteReader *bytes.Buffer, encoding *Encoding) ([]byte, error) {
	reader, err := GetRawReader(byteReader, encoding)
	if err != nil {
		return nil, &ReaderError{
			error: err,
			cause: err,
		}
	}

	return io.ReadAll(reader)
}

func GetRawReader(byteReader *bytes.Buffer, encoding *Encoding) (io.Reader, error) {
	switch encoding.name {
	case Gzip:
		return gzip.NewReader(byteReader)

	case Deflate:
		return flate.NewReader(byteReader), nil

	default:
		return byteReader, nil
	}
}

// Encode data in a []byte based on supplied encoding.
func Encode(data []byte, encoding *Encoding) ([]byte, error) {
	switch encoding.name {
	case Gzip:
		return CompressWithGzip(data)

	case Deflate:
		return CompressWithZlib(data)

	default:
		return data, nil
	}
}

func CompressWithGzip(bodyBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(bodyBytes); err != nil {
		// log.Printf("unable to recompress rewrited body: %v", err)

		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		// log.Printf("unable to close gzip writer: %v", err)

		return nil, err
	}

	return buf.Bytes(), nil
}

func CompressWithZlib(bodyBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	zlibWriter, _ := flate.NewWriter(&buf, flate.DefaultCompression)

	if _, err := zlibWriter.Write(bodyBytes); err != nil {
		// log.Printf("unable to recompress rewrited body: %v", err)

		return nil, err
	}

	if err := zlibWriter.Close(); err != nil {
		// log.Printf("unable to close zlib writer: %v", err)

		return nil, err
	}

	return buf.Bytes(), nil
}

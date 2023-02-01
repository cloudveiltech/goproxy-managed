package main

import (
	"C"
)
import (
	"bytes"
	"io"
	"io/ioutil"

	"compress/flate"
	"compress/gzip"

	"github.com/dsnet/compress/brotli"
)

func decodeResponseCompression(contentEncoding string, body []byte) []byte {
	switch contentEncoding {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewBuffer(body))
		return readReader(reader, err)
	case "br":
		reader, err := brotli.NewReader(bytes.NewBuffer(body), nil)
		if err == nil {
			buf := make([]byte, 1024)
			body = make([]byte, 0)
			defer reader.Close()
			n, _ := reader.Read(buf)
			for n > 0 {
				body = append(body, buf...)
				n, _ = reader.Read(buf)
			}
			return body
		}
	case "deflate":
		reader := flate.NewReader(bytes.NewBuffer(body))
		return readReader(reader, nil)
	}
	return body
}

func readReader(reader io.ReadCloser, err error) []byte {
	if err == nil {
		defer reader.Close()
		body, _ := ioutil.ReadAll(reader)
		return body
	}
	return nil
}

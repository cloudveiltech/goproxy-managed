package main

import (
	"C"
)
import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"compress/flate"
	"compress/gzip"

	"github.com/cloudveiltech/goproxy"
	"github.com/dsnet/compress/brotli"
)

//export ResponseGetStatusCode
func ResponseGetStatusCode(id int64) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}

	return response.StatusCode
}

//export ResponseGetBody
func ResponseGetBody(id int64, res *[]byte) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	if response.Body == nil {
		return false
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	result := buf.Bytes()
	if response.Uncompressed || !adblockMatcher.TestContentTypeIsFiltrable(response.Header.Get("Content-Type")) {
		*res = result
	} else {
		*res = decodeResponseCompression(response.Header.Get("Content-Encoding"), result)
		if *res == nil {
			log.Print("Decoded nil response")
			*res = result
		}
	}

	//since we'd read all body - we need to recreate reader for client here
	response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(result))

	return true
}

//export ResponseGetBodyAsString
func ResponseGetBodyAsString(id int64, res *string) bool {
	var bytes []byte
	if !ResponseGetBody(id, &bytes) {
		return false
	}
	*res = string(bytes[:])

	return true
}

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
	log.Printf("Reader errror %v", err)
	return nil
}

//export ResponseHasBody
func ResponseHasBody(id int64) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	return response.Body != nil && response.ContentLength != 0
}

//export ResponseHeaderExists
func ResponseHeaderExists(id int64, name string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	_, headerExists := response.Header[name]
	return headerExists
}

//export ResponseGetFirstHeader
func ResponseGetFirstHeader(id int64, name string, res *string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	values, headerExists := response.Header[name]
	if !headerExists {
		return false
	}
	*res = values[0]
	return true
}

//export ResponseSetHeader
func ResponseSetHeader(id int64, name string, value string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	response.Header.Set(name, value)
	return true
}

//export ResponseGetHeaders
func ResponseGetHeaders(id int64, keys *string) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	var result strings.Builder
	for key, v := range response.Header {
		for _, value := range v {
			result.WriteString(key + ": " + value + "\r\n")
		}
	}

	*keys = result.String()
	return len(response.Header)
}

//export ResponseGetCertificatesCount
func ResponseGetCertificatesCount(id int64) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	if response.TLS == nil {
		return 0
	}
	return len(response.TLS.PeerCertificates)
}

//export ResponseIsTLSVerified
func ResponseIsTLSVerified(id int64) bool {
	isVerified := isSessionTlsVerified(id)
	return isVerified
}

//export ResponseGetCertificate
func ResponseGetCertificate(id int64, index int32, certData *[]byte) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	if response.TLS == nil {
		return 0
	}

	cert := response.TLS.PeerCertificates[index]

	*certData = cert.Raw

	return 1
}

//export CreateResponse
func CreateResponse(id int64, status int32, contentType string, body string) bool {
	session, exists := sessionMap[id]
	if !exists {
		return false
	}

	session.response = goproxy.NewResponse(session.request, contentType, int(status), body)
	return true
}

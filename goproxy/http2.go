package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/cloudveiltech/goproxy"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

var http2ProxySessionCounter int64

type Http2Handler struct {
	lastHttpResponse map[uint32]*http.Response
	lastHttpRequest  map[uint32]*http.Request
	lastHeadersBlock map[uint32]*http2.HeadersFrameParam
	proxyCtx         map[uint32]*goproxy.ProxyCtx
}

func serveHttp2Filtering(r *http.Request, rawClientTls *tls.Conn, remote *tls.Conn) bool {
	log.Print("Running http2 handler for " + r.URL.String())

	http2Handler := &Http2Handler{
		lastHttpResponse: make(map[uint32]*http.Response),
		lastHeadersBlock: make(map[uint32]*http2.HeadersFrameParam),
		lastHttpRequest:  make(map[uint32]*http.Request),
		proxyCtx:         make(map[uint32]*goproxy.ProxyCtx),
	}
	go func() {
		http2Handler.processHttp2Stream(rawClientTls, remote)
	}()

	return true
}

func (http2Handler *Http2Handler) processHttp2Stream(local *tls.Conn, remote *tls.Conn) {
	const preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	b := make([]byte, len(preface))
	if _, err := io.ReadFull(local, b); err != nil {
		log.Printf("ReadFrame: preface %v", err)
		return
	}
	if string(b) != preface {
		log.Printf("ReadFrame: preface error")
		return
	}
	remote.Write(b)

	http2.VerboseLogs = false
	directFramer := http2.NewFramer(remote, local)
	reverseFramer := http2.NewFramer(local, remote)

	go func() {
		defer remote.Close()
		defer local.Close()
		decoder := hpack.NewDecoder(400096, nil)
		for {
			if !http2Handler.readFrame(reverseFramer, directFramer, decoder, false) {
				return
			}
		}
	}()

	decoder := hpack.NewDecoder(400096, nil)
	for {
		if !http2Handler.readFrame(directFramer, reverseFramer, decoder, true) {
			return
		}
	}
}

func isContentTypeFilterable(contentType string) bool {
	return strings.Contains(contentType, "html") ||
		strings.Contains(contentType, "json")
}

func (http2Handler *Http2Handler) readFrame(directFramer, reverseFramer *http2.Framer, decoder *hpack.Decoder, client bool) bool {
	f, err := directFramer.ReadFrame()
	if err != nil {
		log.Printf("ReadFrame client %v, err: %v", client, err)
		return false
	}

	switch f.Header().Type {
	case http2.FrameData:
		fr := f.(*http2.DataFrame)
		body := fr.Data()

		lastHttpResponse := http2Handler.lastHttpResponse[f.Header().StreamID]
		if lastHttpResponse != nil && !client {
			contentType := lastHttpResponse.Header.Get("Content-Type")
			if isContentTypeFilterable(contentType) {
				putResponseBody(body, lastHttpResponse)
				ctx := http2Handler.proxyCtx[f.Header().StreamID]
				resp := proxy.FilterResponse(lastHttpResponse, ctx)

				if resp != lastHttpResponse { //new response
					writeHeaders(directFramer, &http2.HeadersFrameParam{
						StreamID:      f.Header().StreamID,
						BlockFragment: encodeHeaders(resp),
						EndStream:     false,
						EndHeaders:    true,
						PadLength:     0,
						Priority:      http2.PriorityParam{},
					})
					buf := new(bytes.Buffer)
					buf.ReadFrom(resp.Body)
					directFramer.WriteData(f.Header().StreamID, true, buf.Bytes())
					directFramer.WriteGoAway(f.Header().StreamID, http2.ErrCodeCancel, nil)
					//	delete(http2Handler.lastHttpResponse, f.Header().StreamID)
					//	delete(http2Handler.lastHttpRequest, f.Header().StreamID)
					return false
				}
			}
		}

		header, ok := http2Handler.lastHeadersBlock[f.Header().StreamID]
		if ok {
			header.EndStream = false
			writeHeaders(directFramer, header)
			delete(http2Handler.lastHeadersBlock, f.Header().StreamID)
		}

		directFramer.WriteData(f.Header().StreamID, fr.StreamEnded(), body)
	case http2.FrameHeaders:
		fr := f.(*http2.HeadersFrame)
		decoder.SetMaxStringLength(16 << 20)
		headerFields, headerBlock := decodeAllHeaders(directFramer, fr, decoder)
		if len(headerFields) == 0 {
			log.Printf("Error parsing headers")
		}
		if client {
			request := makeHttpRequest(nil, headerFields)
			var ctx = &goproxy.ProxyCtx{Req: request, Session: atomic.AddInt64(&http2ProxySessionCounter, 1)}
			http2Handler.lastHttpRequest[f.Header().StreamID] = request
			http2Handler.proxyCtx[f.Header().StreamID] = ctx
			_, resp := proxy.FilterRequest(request, ctx)
			log.Printf("FILTERING Request %s", request.RequestURI)
			if resp != nil {
				log.Printf("FILTERING Request blocked %s", request.RequestURI)
				writeHeaders(reverseFramer, &http2.HeadersFrameParam{
					StreamID:      f.Header().StreamID,
					BlockFragment: encodeHeaders(resp),
					EndStream:     false,
					EndHeaders:    true,
					PadLength:     0,
					Priority:      fr.Priority,
				})
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				reverseFramer.WriteData(f.Header().StreamID, true, buf.Bytes())
				reverseFramer.WriteGoAway(f.Header().StreamID, http2.ErrCodeCancel, nil)
				return false
			}
		} else {
			http2Handler.lastHttpResponse[f.Header().StreamID] = makeHttpResponse(nil, headerFields)
			http2Handler.lastHttpResponse[f.Header().StreamID].Request = http2Handler.lastHttpRequest[f.Header().StreamID]
		}

		header := http2.HeadersFrameParam{
			StreamID:      f.Header().StreamID,
			BlockFragment: headerBlock,
			EndStream:     fr.StreamEnded(),
			EndHeaders:    fr.HeadersEnded(),
			PadLength:     0,
			Priority:      fr.Priority,
		}

		if client || fr.StreamEnded() {
			writeHeaders(directFramer, &header)
		} else {
			http2Handler.lastHeadersBlock[f.Header().StreamID] = &header
		}
	case http2.FramePriority:
		fr := f.(*http2.PriorityFrame)
		directFramer.WritePriority(f.Header().StreamID, fr.PriorityParam)
	case http2.FrameRSTStream:
		fr := f.(*http2.RSTStreamFrame)
		directFramer.WriteRSTStream(f.Header().StreamID, fr.ErrCode)
	case http2.FrameSettings:
		fr := f.(*http2.SettingsFrame)
		if fr.IsAck() {
			directFramer.WriteSettingsAck()
		} else {
			params := make([]http2.Setting, 0)
			for i := 0; i < fr.NumSettings(); i++ {
				params = append(params, fr.Setting(i))
			}
			directFramer.WriteSettings(params...)
		}
	case http2.FramePushPromise:
		fr := f.(*http2.PushPromiseFrame)
		directFramer.WritePushPromise(http2.PushPromiseParam{
			StreamID:      f.Header().StreamID,
			PromiseID:     fr.PromiseID,
			BlockFragment: fr.HeaderBlockFragment(),
			EndHeaders:    fr.HeadersEnded(),
			PadLength:     0,
		})
	case http2.FramePing:
		fr := f.(*http2.PingFrame)
		directFramer.WritePing(fr.IsAck(), fr.Data)
	case http2.FrameGoAway:
		fr := f.(*http2.GoAwayFrame)
		directFramer.WriteGoAway(fr.LastStreamID, fr.ErrCode, fr.DebugData())
	case http2.FrameWindowUpdate:
		fr := f.(*http2.WindowUpdateFrame)
		directFramer.WriteWindowUpdate(f.Header().StreamID, fr.Increment)
	case http2.FrameContinuation:
	//	fr := f.(*http2.ContinuationFrame)
	//	directFramer.WriteContinuation(f.Header().StreamID, fr.HeadersEnded(), fr.HeaderBlockFragment())
	default:
		fr := f.(*http2.UnknownFrame)
		directFramer.WriteRawFrame(f.Header().Type, f.Header().Flags, f.Header().StreamID, fr.Payload())
	}

	return true
}

func decodeAllHeaders(framer *http2.Framer, fr *http2.HeadersFrame, decoder *hpack.Decoder) ([]hpack.HeaderField, []byte) {
	buf := new(bytes.Buffer)
	res := make([]hpack.HeaderField, 0)
	decoder.SetEmitEnabled(true)
	decoder.SetMaxStringLength(16 << 20)
	decoder.SetEmitFunc(func(hf hpack.HeaderField) {
		if len(hf.Name) > 0 {
			res = append(res, hf)
		}
	})
	defer decoder.SetEmitFunc(func(hf hpack.HeaderField) {})
	defer decoder.Close()

	_, err := decoder.Write(fr.HeaderBlockFragment())
	if err != nil {
		log.Printf("Error decode %v", err)
	}
	buf.Write(fr.HeaderBlockFragment())
	if fr.HeadersEnded() {
		return res, buf.Bytes()
	}
	for {
		if f, err := framer.ReadFrame(); err != nil {
			break
		} else {
			continuationFrame := f.(*http2.ContinuationFrame) // guaranteed by checkFrameOrder
			decoder.Write(continuationFrame.HeaderBlockFragment())
			_, err = buf.Write(continuationFrame.HeaderBlockFragment())
			if err != nil {
				log.Printf("Error decode %v", err)
			}
			if continuationFrame.HeadersEnded() {
				break
			}
		}
	}
	return res, buf.Bytes()
}

func writeHeaders(framer *http2.Framer, param *http2.HeadersFrameParam) {
	dataToSend := param.BlockFragment
	chunkSize := 15 * 1024
	for i := 0; i < len(dataToSend); i += chunkSize {
		end := i + chunkSize
		headesEnd := false
		if end >= len(dataToSend) {
			end = len(dataToSend)
			headesEnd = true
		}

		if i == 0 {
			framer.WriteHeaders(http2.HeadersFrameParam{
				StreamID:      param.StreamID,
				BlockFragment: dataToSend[i:end],
				EndStream:     headesEnd && param.EndStream,
				EndHeaders:    headesEnd,
				PadLength:     0,
				Priority:      param.Priority,
			})
		} else {
			framer.WriteContinuation(param.StreamID, headesEnd, dataToSend[i:end])
		}
	}
}

func makeHttpRequest(body []byte, header []hpack.HeaderField) *http.Request {
	req := http.Request{}
	req.Proto = "http/2"
	req.ProtoMajor = 2
	req.ProtoMinor = 0

	req.Header = http.Header{}
	scheme := "https"
	authority := ""
	path := ""
	for _, v := range header {
		if !strings.HasPrefix(v.Name, ":") {
			req.Header.Add(v.Name, v.Value)
		} else if v.Name == ":scheme" {
			scheme = v.Value
		} else if v.Name == ":authority" {
			authority = v.Value
		} else if v.Name == ":path" {
			path = v.Value
		} else if v.Name == ":method" {
			req.Method = v.Value
		}
	}
	req.RequestURI = scheme + "://" + authority + path
	req.URL, _ = url.ParseRequestURI(req.RequestURI)
	req.Host = req.URL.Host

	if len(body) == 0 {
		req.Body = http.NoBody
	} else {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
	}
	return &req
}

func makeHttpResponse(body []byte, header []hpack.HeaderField) *http.Response {
	resp := http.Response{}

	resp.Proto = "http/2"
	resp.ProtoMajor = 2
	resp.ProtoMinor = 0
	resp.Header = http.Header{}
	for _, v := range header {
		if !strings.HasPrefix(v.Name, ":") {
			resp.Header.Add(v.Name, v.Value)
		} else if v.Name == ":status" {
			resp.Status = v.Value
			resp.StatusCode, _ = strconv.Atoi(v.Value)
		}
	}
	if len(body) == 0 {
		resp.Body = http.NoBody
		resp.ContentLength = 0
	} else {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
	}
	return &resp
}

func putResponseBody(body []byte, resp *http.Response) {
	if len(body) == 0 {
		resp.Body = http.NoBody
	} else {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
	}
}

func encodeHeadersMetaFrame(fr *http2.MetaHeadersFrame) []byte {
	buf := new(bytes.Buffer)
	encoder := hpack.NewEncoder(buf)
	encoder.SetMaxDynamicTableSize(4096)
	buf.Reset()

	for i := 0; i < len(fr.Fields); i++ {
		encoder.WriteField(fr.Fields[i])
	}
	return buf.Bytes()
}

func encodeHeaders(resp *http.Response) []byte {
	buf := new(bytes.Buffer)
	encoder := hpack.NewEncoder(buf)
	encoder.SetMaxDynamicTableSize(4096)
	buf.Reset()

	writeHeader(encoder, ":status", strconv.Itoa(resp.StatusCode))
	for k, vv := range resp.Header {
		lowKey := strings.ToLower(k)
		for _, v := range vv {
			writeHeader(encoder, lowKey, v)
		}
	}
	return buf.Bytes()
}

func writeHeader(encoder *hpack.Encoder, name, value string) {
	encoder.WriteField(hpack.HeaderField{Name: name, Value: value})
}

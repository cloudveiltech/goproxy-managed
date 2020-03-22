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

	"github.com/cloudveiltech/goproxy"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

type Http2Handler struct {
	lastHttpResponse map[uint32]*http.Response
	lastHttpRequest  map[uint32]*http.Request
	lastHeadersBlock map[uint32]*http2.HeadersFrameParam
	data             map[uint32]*bytes.Buffer
}

func serveHttp2Filtering(r *http.Request, rawClientTls *tls.Conn, remote *tls.Conn) bool {
	log.Print("Running http2 handler for " + r.URL.String())
	ctx := &goproxy.ProxyCtx{Req: r, Session: 1}

	http2Handler := &Http2Handler{
		lastHttpResponse: make(map[uint32]*http.Response),
		lastHeadersBlock: make(map[uint32]*http2.HeadersFrameParam),
		lastHttpRequest:  make(map[uint32]*http.Request),
		data:             make(map[uint32]*bytes.Buffer),
	}
	go func() {
		http2Handler.processHttp2Stream(rawClientTls, remote, ctx)
	}()

	return true
}

/*
func processHttp2Stream1(read *tls.Conn, write *tls.Conn) {
	for {
		n, err := io.Copy(write, read)
		if err != nil {
			return
		}
		if n == 0 {
			return
		}
		time.Sleep(time.Millisecond) //reduce CPU usage due to infinite nonblocking loop
	}
}*/

func (http2Handler *Http2Handler) processHttp2Stream(local *tls.Conn, remote *tls.Conn, ctx *goproxy.ProxyCtx) {
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

	directFramer := http2.NewFramer(remote, local)
	reverseFramer := http2.NewFramer(local, remote)

	//	defer remote.Close()
	//	defer local.Close()
	go func() {
		//		defer remote.Close()
		//	defer local.Close()
		for {
			if !http2Handler.readFrame(directFramer, reverseFramer, ctx, true) {
				return
			}
		}
	}()
	for {
		if !http2Handler.readFrame(reverseFramer, directFramer, ctx, false) {
			return
		}
	}
}

func isContentTypeFilterable(contentType string) bool {
	return strings.Contains(contentType, "html") ||
		strings.Contains(contentType, "json")
}

func (http2Handler *Http2Handler) readFrame(directFramer, reverseFramer *http2.Framer, ctx *goproxy.ProxyCtx, client bool) bool {
	f, err := directFramer.ReadFrame()
	if err != nil {
		log.Printf("ReadFrame client %v, err: %v", client, err)
		return false
	}

	switch f.Header().Type {
	case http2.FrameData:
		fr := f.(*http2.DataFrame)
		body := fr.Data()
		bodyBuf, ok := http2Handler.data[f.Header().StreamID]
		if !ok {
			bodyBuf = new(bytes.Buffer)
			http2Handler.data[f.Header().StreamID] = bodyBuf
		}
		bodyBuf.Write(body)

		lastHttpResponse := http2Handler.lastHttpResponse[f.Header().StreamID]
		if lastHttpResponse != nil && !client && fr.StreamEnded() {
			contentType := lastHttpResponse.Header.Get("Content-Type")
			if isContentTypeFilterable(contentType) {
				body := bodyBuf.Bytes()
				putResponseBody(body, lastHttpResponse)
				resp := proxy.FilterResponse(lastHttpResponse, ctx)

				if resp != lastHttpResponse { //new response
					directFramer.WriteHeaders(http2.HeadersFrameParam{
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

					delete(http2Handler.data, f.Header().StreamID)
					return false
				}
			}

			header, ok := http2Handler.lastHeadersBlock[f.Header().StreamID]
			if ok {
				directFramer.WriteHeaders(*header)
				delete(http2Handler.lastHeadersBlock, f.Header().StreamID)
			}
		}

		if fr.StreamEnded() {
			dataToSend := bodyBuf.Bytes()
			chunkSize := 15 * 1024
			for i := 0; i < len(dataToSend); i += chunkSize {
				end := i + chunkSize
				streamEnd := false
				if end > len(dataToSend) {
					end = len(dataToSend) - 1
					streamEnd = true
				}

				directFramer.WriteData(f.Header().StreamID, streamEnd, dataToSend[i:end])
			}
			delete(http2Handler.data, f.Header().StreamID)
		}
	case http2.FrameHeaders:
		fr := f.(*http2.HeadersFrame)

		decoder := hpack.NewDecoder(204800, nil)
		headerBlock := fr.HeaderBlockFragment()
		hf, err := decoder.DecodeFull(headerBlock)
		if err != nil {
			log.Printf("Decode header err %v", err)
		}

		lastHeader := make(map[string]string)
		for _, h := range hf {
			lastHeader[h.Name] = h.Value
		}
		if client && len(lastHeader) > 0 {
			request := makeHttpRequest(nil, lastHeader)
			http2Handler.lastHttpRequest[f.Header().StreamID] = request
			_, resp := proxy.FilterRequest(request, ctx)
			if resp != nil {
				reverseFramer.WriteHeaders(http2.HeadersFrameParam{
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
		} else if !client && len(lastHeader) > 0 {
			http2Handler.lastHttpResponse[f.Header().StreamID] = makeHttpResponse(nil, lastHeader)
			http2Handler.lastHttpResponse[f.Header().StreamID].Request = http2Handler.lastHttpRequest[f.Header().StreamID]
		}
		if client || len(lastHeader) == 0 || fr.StreamEnded() {
			directFramer.WriteHeaders(http2.HeadersFrameParam{
				StreamID:      f.Header().StreamID,
				BlockFragment: fr.HeaderBlockFragment(),
				EndStream:     fr.StreamEnded(),
				EndHeaders:    fr.HeadersEnded(),
				PadLength:     0,
				Priority:      fr.Priority,
			})
		} else {
			bufferCopy := make([]byte, len(fr.HeaderBlockFragment()))
			copy(bufferCopy, fr.HeaderBlockFragment())
			header := http2.HeadersFrameParam{
				StreamID:      f.Header().StreamID,
				BlockFragment: bufferCopy,
				EndStream:     fr.StreamEnded(),
				EndHeaders:    fr.HeadersEnded(),
				PadLength:     0,
				Priority:      fr.Priority,
			}

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
		fr := f.(*http2.ContinuationFrame)
		directFramer.WriteContinuation(f.Header().StreamID, fr.HeadersEnded(), fr.HeaderBlockFragment())
	default:
		fr := f.(*http2.UnknownFrame)
		directFramer.WriteRawFrame(f.Header().Type, f.Header().Flags, f.Header().StreamID, fr.Payload())
	}

	return true
}

func makeHttpRequest(body []byte, header map[string]string) *http.Request {
	req := http.Request{}
	req.Method = header[":method"]
	req.Proto = "http/2"
	req.RequestURI = header[":scheme"] + "://" + header[":authority"] + header[":path"]
	req.ProtoMajor = 2
	req.ProtoMinor = 0
	req.URL, _ = url.ParseRequestURI(req.RequestURI)
	req.Host = req.URL.Host

	req.Header = http.Header{}
	for k, v := range header {
		if !strings.HasPrefix(k, ":") {
			req.Header.Add(k, v)
		}
	}

	if len(body) == 0 {
		req.Body = http.NoBody
	} else {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
	}
	return &req
}

func makeHttpResponse(body []byte, header map[string]string) *http.Response {
	resp := http.Response{}

	resp.Proto = "http/2"
	resp.Status = header[":status"] //??
	resp.StatusCode, _ = strconv.Atoi(header[":status"])
	resp.ProtoMajor = 2
	resp.ProtoMinor = 0
	resp.Header = http.Header{}
	for k, v := range header {
		if !strings.HasPrefix(k, ":") {
			resp.Header.Add(k, v)
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

func encodeHeaders(resp *http.Response) []byte {
	buf := new(bytes.Buffer)
	encoder := hpack.NewEncoder(buf)
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

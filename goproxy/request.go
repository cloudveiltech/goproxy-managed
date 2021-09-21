package main

import (
	"C"
	"bytes"
)
import (
	"io/ioutil"
	"strings"
)

//export RequestGetUrl
func RequestGetUrl(id int64, result *string) bool {
	request := getSessionRequest(id)

	if request == nil {
		return false
	}

	if request.URL == nil {
		return false
	}

	*result = request.URL.String()
	return len(*result) > 0
}

//export RequestGetBody
func RequestGetBody(id int64, res *[]byte) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	if request.Body == nil {
		return false
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)

	*res = buf.Bytes()

	//since we'd read all body - we need to recreate reader for client here
	request.Body.Close()
	request.Body = ioutil.NopCloser(bytes.NewBuffer(*res))

	return true
}

//export RequestGetBodyAsString
func RequestGetBodyAsString(id int64, res *string) bool {
	var bytes []byte
	if !RequestGetBody(id, &bytes) {
		return false
	}
	*res = string(bytes[:])

	return true
}

//export RequestHasBody
func RequestHasBody(id int64) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	return request.Body != nil && request.ContentLength != 0
}

//export RequestHeaderExists
func RequestHeaderExists(id int64, name string) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	// for k := range request.Header {
	// 	fmt.Fprintf(os.Stderr, "key[%s] value[%s]\n", k, request.Header[k])
	// }

	_, headerExists := request.Header[name]
	return headerExists
}

//export RequestGetFirstHeader
func RequestGetFirstHeader(id int64, name string, res *string) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	values, headerExists := request.Header[name]
	if !headerExists {
		return false
	}
	*res = values[0]
	return true
}

//export RequestSetHeader
func RequestSetHeader(id int64, name string, value string) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	request.Header.Set(name, value)
	return true
}

//export RequestGetHeaders
func RequestGetHeaders(id int64, keys *string) int {
	request := getSessionRequest(id)
	if request == nil {
		return 0
	}
	var result strings.Builder
	for key, v := range request.Header {
		for _, value := range v {
			result.WriteString(key + ": " + value + "\r\n")
		}
	}

	*keys = result.String()

	return len(request.Header)
}

func HostPathForceSafeSearch(host, path string) string {
	// enforce Google safe-search
	if strings.Contains(host, "google.com") && strings.Contains(path, "/search?") && !strings.Contains(path, "safe=active") {
		return strings.Replace(path+"&safe=active", "&safe=images", "", -1)
		// enforce Bing safe-search
	} else if strings.Contains(host, "bing.com") && strings.Contains(path, "/search?") && !strings.Contains(path, "adlt=strict") {
		return path + "&adlt=strict"
		// enforce Yahoo safe-search
	} else if strings.Contains(host, "yahoo.com") && strings.Contains(path, "/search?") && !strings.Contains(path, "&vm=r") {
		return path + "&vm=r"
	}
	return path
}

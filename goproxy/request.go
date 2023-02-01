package main

import (
	"C"
)
import (
	"strings"
)

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

func CookiePatchSafeSearch(host, cookieValue string) string {
	if strings.Contains(host, "vimeo") {
		cookieParts := strings.Split(cookieValue, ";")
		newCookieParts := make([]string, 0)
		for _, cookie := range cookieParts {
			if !strings.Contains(cookie, "content_rating") {
				newCookieParts = append(newCookieParts, cookie)
			}
		}

		newCookieParts = append(newCookieParts, "content_rating=7")

		return strings.Join(newCookieParts, ";")
	}
	return cookieValue
}

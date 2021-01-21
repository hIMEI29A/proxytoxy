// Copyright 2020  himei@tuta.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/hIMEI29A/proxytoxy/proxy"
)

var (
	prxPattern     = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\:\d{1,5}\#[A-Z]{2}`)
	ipPattern      = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	prPortPattern  = regexp.MustCompile(`\:\d{2,5}`)
	countryPattern = regexp.MustCompile(`[A-Z]{2}`)
)

func extractProxyStrings(proxyStrings string) []string {
	var ss []string

	sb := prxPattern.FindAll([]byte(proxyStrings), -1)

	for i := range sb {
		ss = append(ss, string(sb[i]))
	}

	return ss
}

func extractIp(proxyString string) string {
	return ipPattern.FindString(proxyString)
}

func extractPort(proxyString string) string {
	s := prPortPattern.FindString(proxyString)
	return strings.TrimPrefix(s, ":")
}

func extractCountry(proxyString string) string {
	return countryPattern.FindString(proxyString)
}

func myproxyFunc(url string, c *colly.Collector) []*proxy.Proxy {
	var (
		proxies []*proxy.Proxy
		t       string
	)

	c.OnHTML(".col-lg-12", func(e *colly.HTMLElement) {
		if strings.Contains(e.ChildText("h2"), "Socks5 Proxy") {
			t = "socks5"
		}

		if strings.Contains(e.ChildText("h2"), "Proxy List #") {
			t = "http"
		}
	})

	c.OnHTML(".list", func(e *colly.HTMLElement) {
		var pro = &proxy.Proxy{}
		proxyStrings := extractProxyStrings(e.Text)

		for i := range proxyStrings {
			ip := extractIp(proxyStrings[i])
			port := extractPort(proxyStrings[i])
			unparsedCountry := extractCountry(proxyStrings[i])
			country := parseCountry(unparsedCountry)

			if country != unknownCountry {
				pro = proxy.NewProxy(ip, port, country, "socks5", proxy.Elite)
				proxies = append(proxies, pro)
			}
		}
	})

	c.OnHTML(".list-group", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Attr("href")
			if strings.Contains(link, "free-proxy-list-") {
				c.Visit(el.Request.AbsoluteURL(link))
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)

	return proxies
}

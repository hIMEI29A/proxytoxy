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
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/hIMEI29A/proxytoxy/proxy"
)

func proxylistplusFunc(url string, c *colly.Collector) []*proxy.Proxy {
	var (
		proxies []*proxy.Proxy
		tt      string
		fields  = make([][]string, 0)
	)

	c.OnHTML("h3", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Socks Proxy List - Verified") {
			tt = "socks5"
		}

		if strings.Contains(e.Text, "Free  Proxy List - verified") {
			tt = "http"
		}
	})

	c.OnHTML(".bg", func(e *colly.HTMLElement) {
		var pro = &proxy.Proxy{}
		e.ForEach(".cells", func(_ int, el *colly.HTMLElement) {
			f := make([]string, 0)

			el.ForEach("td", func(_ int, le *colly.HTMLElement) {
				f = append(f, le.Text)
			})

			fields = append(fields, f)

		})

		for i := 0; i < len(fields); i++ {
			ip := fields[i][1]
			port := fields[i][2]

			if strings.Contains(port, "\n") {
				psp := strings.Split(port, "\n")
				port = psp[2]
			}

			t := tt
			r := proxy.Anonymous
			unparsedCountry := fields[i][4]
			country := parseCountry(unparsedCountry)

			if country != unknownCountry {
				pro = proxy.NewProxy(ip, port, country, t, r)
				proxies = append(proxies, pro)
			}
		}
	})

	c.OnHTML(".cells", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Attr("href")

			if strings.Contains(link, "Fresh-HTTP-Proxy-List-") {
				c.Visit(el.Request.AbsoluteURL(link))
			}

		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
	c.Visit("https://list.proxylistplus.com/Socks-List-2")

	return proxies
}

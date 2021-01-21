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

func nntimeFunc(url string, c *colly.Collector) []*proxy.Proxy {
	var (
		ps      string
		proxies []*proxy.Proxy

		fn = func(e *colly.HTMLElement) {
			var pro = &proxy.Proxy{}

			p := e.ChildText("script")
			port := decodePort(ps, p)

			fields := e.ChildTexts("td")

			ip := strings.Split(fields[1], "d")[0]

			t := "http"
			r := proxy.Anonymous
			unparsedCountry := strings.Split(fields[4], " ")[0]
			country := parseCountry(unparsedCountry)

			if country != unknownCountry {
				pro = proxy.NewProxy(ip, port, country, t, r)
				proxies = append(proxies, pro)
			}
		}
	)

	c.OnHTML("head", func(e *colly.HTMLElement) {
		ps = e.ChildTexts("script")[1]
	})

	c.OnHTML(".odd", fn)

	c.OnHTML(".even", fn)

	c.OnHTML("#navigation", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := el.Attr("href")
			c.Visit(el.Request.AbsoluteURL(link))
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)

	return proxies
}

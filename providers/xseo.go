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

func xseoFunc(url string, c *colly.Collector) []*proxy.Proxy {
	var (
		ps         string
		proxies    []*proxy.Proxy
		proxyRange int

		titleSlice = []string{
			"IP адрес и порт proxy",
			"Хостнейм",
			"Тип",
			"*Анонимность",
			"Страна",
			"Дата проверки",
		}

		fn = func(e *colly.HTMLElement) {
			var pro = &proxy.Proxy{}

			p := e.ChildText("script")
			port := decodePort(ps, p)

			if !compareSlices(titleSlice, e.ChildTexts("td")) {
				fields := e.ChildTexts("td")

				ip := strings.Split(fields[0], ":")[0]
				t := strings.ToLower(fields[2])
				r := fields[3]

				if r == "да" {
					proxyRange = proxy.Anonymous
				}

				if r == "да+" {
					proxyRange = proxy.Elite
				}

				unparsedCountry := strings.Split(fields[4], " ")[0]
				country := parseCountry(unparsedCountry)

				if country != unknownCountry {
					pro = proxy.NewProxy(ip, port, country, t, proxyRange)
					proxies = append(proxies, pro)
				}
			}
		}
	)

	requestData := map[string]string{
		"action": "/proxylist",
		"method": "post",
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		ps = e.ChildTexts("script")[0]

	})

	c.OnHTML(".cls81", fn)

	c.OnHTML(".cls8", fn)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Post(url, requestData)

	return proxies
}

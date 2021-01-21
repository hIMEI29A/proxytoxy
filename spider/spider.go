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

package spider

import (
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly/v2"
	collyProxy "github.com/gocolly/colly/v2/proxy"
	"github.com/hIMEI29A/proxytoxy/providers"
	"github.com/hIMEI29A/proxytoxy/proxy"
)

type Spider struct {
	providers.Provider
}

func NewSpider() *Spider {
	return &Spider{}
}

func (s *Spider) setCollector(collectorProxies []string) *colly.Collector {
	c := colly.NewCollector()

	if collectorProxies != nil {
		ps, err := collyProxy.RoundRobinProxySwitcher(collectorProxies...)

		if err != nil {
			log.Fatal(err)
		}

		c.SetProxyFunc(ps)
	}

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       1 * time.Second,
		RandomDelay: 2 * time.Second,
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	return c
}

func (s *Spider) CollectAll(cp []string) []*proxy.Proxy {
	var (
		p [][]*proxy.Proxy
	)

	c := s.setCollector(cp)

	funcs := s.GetProviderFuncs()

	for k, v := range funcs {
		cc := c.Clone()

		pr := v(k, cc)

		p = append(p, pr)
	}

	l := 0

	for _, v := range p {
		l += len(v)
	}

	proxies := make([]*proxy.Proxy, 0, l)
	control := make(map[string]bool)

	for _, v := range p {
		for _, vv := range v {
			if !control[vv.GetFullAddr()] {
				proxies = append(proxies, vv)
				control[vv.GetFullAddr()] = true
			}
		}
	}

	fmt.Printf("%d proxies collected. We must to check them before use.\n", len(proxies))

	return proxies
}

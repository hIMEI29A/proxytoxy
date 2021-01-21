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
	"regexp"

	"github.com/biter777/countries"
	"github.com/gocolly/colly/v2"
	"github.com/hIMEI29A/proxytoxy/proxy"
)

const unknownCountry = "unknown"

var providersUrls = []string{
	"https://xseo.in/proxylist",
	"http://nntime.com",
	"https://www.my-proxy.com/free-socks-5-proxy.html",
	"https://list.proxylistplus.com/Socks-List-1",
	//"https://www.proxy-list.download/SOCKS5",
	//"https://www.xroxy.com/proxylist.htm",
	//"http://www.proxz.com/proxy_list_high_anonymous_0_ext.html",
	//"http://www.proxylists.net/",
	//"https://www.socks-proxy.net/",
	//"https://free-proxy-list.net",
	//"https://us-proxy.org/",
}

var providersFuncs = []ProviderFunc{
	xseoFunc,
	nntimeFunc,
	myproxyFunc,
	proxylistplusFunc,
}

var (
	countryAlpha2Pattern = regexp.MustCompile(`\b([A-Z]{2})\b`)
)

type ProviderFunc func(string, *colly.Collector) []*proxy.Proxy

var (
	portScriptPattern = regexp.MustCompile(`(\w{1}\={1}\d{1})`)
	portPattern       = regexp.MustCompile(`(\+{1}\w{1})`)
)

func parseCountry(unparsed string) string {
	var country string

	if countryAlpha2Pattern.MatchString(unparsed) {
		country = unparsed
	} else {
		code := countries.ByName(unparsed)

		if code == countries.Unknown {
			country = unknownCountry
		} else {
			country = code.Alpha2()
		}
	}

	return country
}

func decodePortScript(ps string) map[string]string {
	mp := make(map[string]string, 10)

	s := portScriptPattern.FindAll([]byte(ps), -1)

	for i := range s {
		mp[string(s[i][0])] = string(s[i][2])
	}

	return mp
}

func decodePort(ps, p string) string {
	var (
		port string
		mp   = decodePortScript(ps)
	)

	s := portPattern.FindAll([]byte(p), -1)

	for i := range s {
		m := string(s[i][1])
		port = port + mp[m]
	}

	return port
}

func compareSlices(a, b []string) bool {
	check := true

	if len(a) == len(b) {
		for i, v := range a {
			if v != b[i] {
				check = false
			}
		}
	}

	return check
}

type Provider struct{}

func NewProvider() Provider {
	return Provider{}
}

func (p Provider) GetProviderFuncs() map[string]ProviderFunc {
	mp := make(map[string]ProviderFunc, 0)

	for k, v := range providersUrls {
		mp[v] = providersFuncs[k]
	}

	return mp
}

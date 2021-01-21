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

package proxy

import (
	//"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	ident = "http://ident.me"
)

const (
	Anonymous = iota
	Elite
	NotAnonymous
)

const (
	ipCheckErr  = "My ip cannot be checked"
	clientErr   = "Error of proxy's address parsing"
	responseErr = "Proxy host not responding"
	rangeErr    = "Proxy host did not change my ip"
)

var proxyRanges = []string{
	"Anonymous",
	"Elite",
	"NotAnonymous",
}

func stringRange(r int) string {
	return proxyRanges[r]
}

func getIp(c *http.Client) (string, error) {
	var (
		ip string
	)

	res, err := c.Get(ident)

	if err != nil {
		return ip, err
	}

	i, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return ip, err
	}

	return string(i), err
}

func getSocksClient(fullAddr string) (*http.Client, error) {
	var err error

	socks, err := proxy.SOCKS5("tcp", fullAddr, nil, &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 15 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Transport: &http.Transport{
			Dial:                socks.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	return c, err
}

func getHttpClient(fullAddr string) (*http.Client, error) {
	var err error

	proxyAddr, err := url.Parse("http://" + fullAddr)

	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyAddr),
		},

		Timeout: 15 * time.Second,
	}

	return c, err
}

type ProxyCheckError struct {
	ResponseErr string // proxy is dead
	IpCheckErr  string // ident.me is dead
	RangeErr    string // proxy is not anonymous
	ClientErr   string // error of parsing proxy address
}

func (pe ProxyCheckError) Error() string {
	var e string

	switch {
	case pe.ResponseErr != "":
		e = e + pe.ResponseErr

	case pe.IpCheckErr != "":
		e = e + pe.IpCheckErr

	case pe.RangeErr != "":
		e = e + pe.RangeErr

	case pe.ClientErr != "":
		e = e + pe.ClientErr
	}

	return e
}

type Proxy struct {
	Addr    string
	Port    string
	Country string
	Type    string
	Range   int
}

func NewProxy(addr, port, country, t string, r int) *Proxy {
	return &Proxy{
		Addr:    addr,
		Port:    port,
		Country: country,
		Type:    t,
		Range:   r,
	}
}

func (p *Proxy) GetFullAddr() string {
	return p.Addr + ":" + p.Port
}

func (p *Proxy) check() error {
	var (
		err = ProxyCheckError{}
		ip  string
	)

	addr := p.GetFullAddr()

	ip, identErr := getIp(http.DefaultClient)

	if identErr != nil {
		err.IpCheckErr = fmt.Sprintf("%s: %s error ", ipCheckErr, ident)
		return err
	}

	switch p.Type {
	case "socks5":
		socksClient, clErr := getSocksClient(addr)

		if clErr != nil {
			err.ClientErr = fmt.Sprintf(clientErr)
			return err
		}

		ipThroughProxy, socksErr := getIp(socksClient)

		if socksErr != nil {
			err.ResponseErr = fmt.Sprintf("%s: %s ", responseErr, addr)
			return err
		}

		if ipThroughProxy == ip {
			err.RangeErr = fmt.Sprintf("%s: %s ", rangeErr, addr)
			return err
		}

	case "http":
		httpClient, clErr := getHttpClient(addr)

		if clErr != nil {
			err.ClientErr = fmt.Sprintf(clientErr)
			return err
		}

		ipThroughProxy, httpErr := getIp(httpClient)

		if httpErr != nil {
			err.ResponseErr = fmt.Sprintf("%s: %s ", responseErr, addr)
			return err
		}

		if ipThroughProxy == ip {
			err.RangeErr = fmt.Sprintf("%s: %s ", rangeErr, addr)
			return err
		}
	}

	return err
}

func (p *Proxy) String() string {
	return fmt.Sprintf(
		"%s:%s %s %s %s",
		p.Addr,
		p.Port,
		p.Country,
		p.Type,
		stringRange(p.Range),
	)
}

func CheckProxies(proxies []*Proxy, anonFlag bool) []*Proxy {
	var (
		checked    = make(map[string]bool, 0)
		newProxies = make([]*Proxy, 0)
		mu         = &sync.Mutex{}
	)

	chanPr := make(chan map[*Proxy]error, 10)

	fmt.Printf("Waiting for %d proxies to be checked.\n", len(proxies))

	for _, p := range proxies {
		if !checked[p.Addr] {
			go func(pr *Proxy, ch chan map[*Proxy]error) {
				mp := make(map[*Proxy]error, 0)
				fmt.Println(pr.GetFullAddr(), "in check")
				err := pr.check()
				fmt.Println(pr.GetFullAddr(), "checked")
				time.Sleep(10 * time.Millisecond)

				mp[pr] = err
				ch <- mp
			}(p, chanPr)
		}

		mu.Lock()
		checked[p.Addr] = true
		mu.Unlock()
	}

	count := len(proxies) - 1

	for count >= 0 {

		select {
		case mp := <-chanPr:
			fmt.Println("recieved", count)
			for k, v := range mp {
				time.Sleep(10 * time.Millisecond)
				ve := v.(ProxyCheckError)
				if ve.ResponseErr == "" && ve.RangeErr != "" {
					k.Range = NotAnonymous

					if anonFlag {
						mu.Lock()
						newProxies = append(newProxies, k)
						mu.Unlock()
					}
				}

				if ve.ResponseErr == "" {
					mu.Lock()
					newProxies = append(newProxies, k)
					mu.Unlock()
				}
			}

			count--
		}
	}

	fmt.Println("All proxied checked. Done.")

	return newProxies
}

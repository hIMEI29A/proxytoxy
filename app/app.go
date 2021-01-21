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

package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hIMEI29A/proxytoxy/proxy"
	"github.com/hIMEI29A/proxytoxy/spider"
)

func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func readFromFile(filepath string) []string {
	var proxies []string

	file, err := os.Open(filepath)
	errFatal(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	return proxies
}

func writeToFile(proxies []*proxy.Proxy, t *Task) {
	var s string

	file, err := os.OpenFile(t.ToFile, os.O_RDWR|os.O_CREATE, 0666)
	errFatal(err)

	defer file.Close()

	for _, v := range proxies {
		if t.Short {
			s = v.GetFullAddr()
		} else {
			s = v.String()
		}

		file.WriteString(s + "\n")
	}
}

type App struct {
	Cli    *Cli
	Spider *spider.Spider
}

func NewApp() *App {
	return &App{
		Cli:    NewCli(),
		Spider: spider.NewSpider(),
	}
}

func (a *App) output(proxies []*proxy.Proxy, w io.Writer, t *Task) {
	var s string

	for _, v := range proxies {
		if t.Short {
			s = v.GetFullAddr()
		} else {
			s = v.String()
		}

		_, err := fmt.Fprintf(w, "%s\n", s)
		errFatal(err)
	}
}

func (a *App) outputStd(proxies []*proxy.Proxy, t *Task) {
	var s string

	for _, v := range proxies {
		if t.Short {
			s = v.GetFullAddr()
		} else {
			s = v.String()
		}

		fmt.Println(s)
	}
}

func (a *App) DoAll(t *Task) []*proxy.Proxy {
	var (
		proxies, useful []*proxy.Proxy
	)

	if t.ProxyFile != "" {
		t.Proxies = readFromFile(t.ProxyFile)
	}

	pr := a.Spider.CollectAll(t.Proxies)

	for _, v := range pr {
		if v.Country == t.Country && v.Type == t.Type {
			useful = append(useful, v)
		}
	}

	checked := proxy.CheckProxies(useful, t.Anon)

	if len(checked) <= t.Number {
		proxies = checked
	} else {
		for i := 0; i < t.Number; i++ {
			proxies = append(proxies, checked[i])
		}
	}

	a.outputStd(proxies, t)

	if t.ToFile != "" {
		writeToFile(proxies, t)
	}

	return proxies
}

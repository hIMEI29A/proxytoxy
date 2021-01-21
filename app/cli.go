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
	"errors"
	"fmt"

	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	helpPattern = regexp.MustCompile(
		`((-h)|(--help))(?: )?`,
	)

	anonPattern = regexp.MustCompile(
		`((-a)|(--anon))(?: )?`,
	)

	shortPattern = regexp.MustCompile(
		`((-a)|(--anon))(?: )?`,
	)

	numberPattern = regexp.MustCompile(
		`((-n)|(--number))(?: )?(?P<number>\b(\d+)\b)(?: )?`,
	)

	typePattern = regexp.MustCompile(
		`((-t)|(--type))(?: )?(?P<type>(socks5)|(http))(?: )?`,
	)

	countryPattern = regexp.MustCompile(
		`((-c)|(--country))(?: )?(?P<country>[A-Z]{2})(?: )?`,
	)

	spiderProxyPattern = regexp.MustCompile(
		`((-p)|(--proxies))(?: )?(?P<proxy>(socks5:\/\/(\d{1,3}\.?){4}:\d{2,5}(?:, )?)+)`,
	)

	proxyFromFilePattern = regexp.MustCompile(
		`((-P)|(--proxy-file))(?: )?(?P<filepath>((\.{2}\/{1})+|((\.{1}\/{1})?)|(\/{1}))((.+\/{1})*)(.)+(\.{1}.+)?)(?: )?`,
	)

	writeToFilePattern = regexp.MustCompile(
		`((-f)|(--file))(?: )?(?P<filepath>((\.{2}\/{1})+|((\.{1}\/{1})?)|(\/{1}))((.+\/{1})*)(.)+(\.{1}.+)?)(?: )?`,
	)
)

type Task struct {
	Type       string // type
	Country    string
	Number     int
	MaxWorkers int
	Proxies    []string
	ProxyFile  string
	ToFile     string
	Anon       bool
	Short      bool
}

const (
	argMissingErr = "Argument is missing"
	numberErr     = "Argument must be a number"
	stringErr     = "Argument must be a string"
	proxiesError  = "--proxies and --proxy-file: only one option allowed"
)

var usage = `
proxytoxy - fast collector of free proxies

Usage: proxytoxy [OPTIONS] [ARGS]

Options:

-h | --help    read this message
		
-c [STRING] | --country [STRING]      The proxy's country. Required.
-n [NUM]    | --number  [NUM]         The number of proxies. Required.
-t [STRING] | --type    [STRING]      The proxy's type. Required.
Allowed types: "socks5", "http".

-p [ARGS...] | --proxies [ARGS...]    One or more socks5 proxies separated by commas
for crawling sites of providers. Format of proxies: "socks5://IP:PORT". 
Not required option.

-P [PATH]    | --proxy-file           Get proxies for crawler from file.
This option can't be set with "-p" option together.

-a           | --anon                 If during the proxy check it turns out that 
the proxy does not hide the IP, such a proxy WILL BE included in the 
app output anyway.

-s           | --short                  Print proxy address and port only  
`

func aToi(s string) int {
	n, err := strconv.Atoi(s)
	errFatal(err)

	return n
}

func help() {
	fmt.Println(usage)
}

type Cli struct{}

func NewCli() *Cli {
	return &Cli{}
}

func (cli *Cli) argsToString(args []string) string {
	return strings.Join(args, " ")
}

func (cli *Cli) helpOption(args []string) bool {
	argString := cli.argsToString(args)

	return helpPattern.MatchString(argString)
}

func (cli *Cli) anonOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if anonPattern.MatchString(argString) {
		t.Anon = true
	}

	return t
}

func (cli *Cli) shortOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if shortPattern.MatchString(argString) {
		t.Short = true
	}

	return t
}

func (cli *Cli) numberOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if !numberPattern.MatchString(argString) {
		err := errors.New(fmt.Sprintf("%s: %s\n", argMissingErr, "-n"))
		help()
		errFatal(err)
	}

	argMatches := numberPattern.FindStringSubmatch(argString)
	i := numberPattern.SubexpIndex("number")

	t.Number = aToi(argMatches[i])

	return t
}

func (cli *Cli) typeOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if !typePattern.MatchString(argString) {
		err := errors.New(fmt.Sprintf("%s: %s\n", argMissingErr, "-t"))
		help()
		errFatal(err)
	}

	argMatches := typePattern.FindStringSubmatch(argString)
	i := typePattern.SubexpIndex("type")

	t.Type = argMatches[i]

	return t
}

func (cli *Cli) countryOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if !countryPattern.MatchString(argString) {
		err := errors.New(fmt.Sprintf("%s: %s\n", argMissingErr, "-c"))
		help()
		errFatal(err)
	}

	argMatches := countryPattern.FindStringSubmatch(argString)
	i := countryPattern.SubexpIndex("country")

	t.Country = argMatches[i]

	return t
}

/*
func (cli *Cli) workersOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	i := 0
	argMatches := workersPattern.FindStringSubmatch(argString)

	if argMatches != nil {
		i = workersPattern.SubexpIndex("workers")

		if i != -1 {
			t.MaxWorkers = aToi(argMatches[i])
		}
	}

	return t
}
*/
func (cli *Cli) toFileOption(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	i := 0
	argMatches := writeToFilePattern.FindStringSubmatch(argString)

	if argMatches != nil {
		i = writeToFilePattern.SubexpIndex("filepath")

		if i != -1 {
			t.ToFile = argMatches[i]
		}
	}

	return t
}

func (cli *Cli) proxiesOption(args []string, t *Task) *Task {
	var pr []string
	argString := cli.argsToString(args)

	if spiderProxyPattern.MatchString(argString) &&
		proxyFromFilePattern.MatchString(argString) {
		err := errors.New(fmt.Sprintf(proxiesError))
		help()
		errFatal(err)
	}

	i := 0
	argMatches := spiderProxyPattern.FindStringSubmatch(argString)

	if argMatches != nil {
		i = spiderProxyPattern.SubexpIndex("proxy")

		if i != -1 {
			pr = strings.Split(argMatches[i], ", ")
		}
	}

	t.Proxies = pr

	return t
}

func (cli *Cli) proxiesFromFile(args []string, t *Task) *Task {
	argString := cli.argsToString(args)

	if spiderProxyPattern.MatchString(argString) &&
		proxyFromFilePattern.MatchString(argString) {
		err := errors.New(fmt.Sprintf(proxiesError))
		help()
		errFatal(err)
	}

	i := 0
	argMatches := proxyFromFilePattern.FindStringSubmatch(argString)

	if argMatches != nil {
		i = proxyFromFilePattern.SubexpIndex("filepath")

		if i != -1 {
			t.ProxyFile = argMatches[i]
		}
	}

	return t
}

func (cli *Cli) Parse(args []string) *Task {
	var (
		task = &Task{}
	)

	// help
	if cli.helpOption(args) {
		help()
		os.Exit(0)
	}

	task = cli.numberOption(args, task)
	task = cli.countryOption(args, task)
	task = cli.typeOption(args, task)
	task = cli.proxiesOption(args, task)
	task = cli.anonOption(args, task)
	task = cli.shortOption(args, task)

	return task
}

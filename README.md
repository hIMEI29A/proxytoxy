# proxytoxy

Proxy grabber. 

## About

Takes proxies from 

[https://xseo.in/proxylist](https://xseo.in/proxylist)

[http://nntime.com](http://nntime.com)

[https://www.my-proxy.com](https://www.my-proxy.com/)

[https://list.proxylistplus.com](https://list.proxylistplus.com/)

### Project status

Work in progress

### Version

v0.1.0-alpha

## Usage

	Usage: proxytoxy [OPTIONS] [ARGS]

	Options:

	-h           | --help                  read this message
		
	-c [STRING]  | --country [STRING]      The proxy's country. Required.
	-n [NUM]     | --number  [NUM]         The number of proxies. Required.
	-t [STRING]  | --type    [STRING]      The proxy's type. Required.
	Allowed types: "socks5", "http".

	-p [ARGS...] | --proxies [ARGS...]    One or more socks5 proxies separated by commas for crawling sites of providers. Format of proxies: "socks5://IP:PORT". Not required option.

	-P [PATH]    | --proxy-file           Get proxies for crawler from file.
	This option can't be set with "-p" option together.

	-a           | --anon                 If during the proxy check it turns out that 
	the proxy does not hide the IP, such a proxy WILL BE included in the 
	app output anyway.

	-s           | --short                Print proxy address and port only  
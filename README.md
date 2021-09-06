# ss-check

A command line tool to check the usability of shadowsocks.

Sometimes shadowsocks server can ping, but `ssserver` may not working at all, for this situation, we can use `ss-check` to check which config is working(By testing `curl http://www.google.com`).

**[中文文档](https://github.com/eleven26/ss-check/blob/master/README-CN.md)**


## Requirements

* macOS > 10.12

* [ShadowsocksX-NG-R8](https://github.com/paradiseduo/ShadowsocksX-NG-R8)


## Installation

### Go

```
go get github.com/eleven26/ss-check 
```

> Remember add `$GOPATH/bin` to your `$PATH` environment variable.

Or

### Curl

```
curl https://raw.githubusercontent.com/eleven26/ss-check/master/install_update_linux.sh | bash
```


## Usage

* Open your `ShadowsocksX-NG-R8`

* Select `All Server To Json...`, select the location to save shadowsocks configuration(JSON format).

* Run the command below:

```
ss-check -c /path/to/config.json
```

`/path/to/config.json` is the location you export from [ShadowsocksX-NG-R8](https://github.com/paradiseduo/ShadowsocksX-NG-R8)

* You can also specify url to test:

```
ss-check -c /path/to/config.json -u www.google.com
```


## How ShadowsocksX-NG work?

![ss-proxy](https://github.com/eleven26/ss-check/blob/master/ss-proxy.png)

# ss-check

一个检查 Shadowsocks 可用性的命令行工具。

有时候 Shadowsocks 服务器可以 ping 得通，但这并不意味着服务器上的 `ssserver` 进程正常运行，对于这种情况，我们可以使用 `ss-check` 来检查哪些配置文件可以正常工作（通过使用 `curl http://www.google.com` 命令）。


## 系统要求

* macOS > 10.12

* [ShadowsocksX-NG-R8](https://github.com/paradiseduo/ShadowsocksX-NG-R8)


## 安装

### Go

```
go get github.com/eleven26/ss-check 
```

> 记得把 `$GOPATH/bin` 加到你的 `$PATH` 环境变量中。

或

### Curl

```
curl https://raw.githubusercontent.com/eleven26/ss-check/master/install_update_linux.sh | bash
```


## 用法

* 打开 `ShadowsocksX-NG-R8`

* 选择 `All Server To Json...`, 然后选择你要保存配置文件的位置（这个配置文件会是一个 JSON 格式）.

* 运行下面的命令:

```
ss-check -c /path/to/config.json
```

* 你也可以指定一个用以测试的域名:

```
ss-check -c /path/to/config.json -u www.google.com
```


`/path/to/config.json` 是你从 [ShadowsocksX-NG-R8](https://github.com/paradiseduo/ShadowsocksX-NG-R8) 导出的文件绝对路径。


## ShadowsocksX-NG 是如何工作的?

![ss-proxy](https://github.com/eleven26/ss-check/blob/master/ss-proxy.png)

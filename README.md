## go-tun2ssh-route

使用go-tun2socks 并注册SshTcpHandler 增加路由表
得到使用ssh转发来自路由数据的能力

```bash
go build go-tun2ssh-route/cmd/tun2ssh-route/main.go
./main -config config_path
```

```yaml
#tun设备占用的ip段
tun-addr: 10.255.123.0
#ssh链接信息
ssh:
  host: 127.0.0.1
  port: 22
  user-name: root
  #  auth-type 可选 password | identity 为identity时, passphrase可选
  auth-type: password
  password: password
  passphrase: passphrase
#  路由表列表
route-arr:
  #  路由地址以及掩码
  - addr: 172.10.0.0
    mask: 255.255.0.0
  - addr: 172.11.0.0
    mask: 255.255.0.0
```
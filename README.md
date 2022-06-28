# TUN
简易隧道工具

## 使用场景
比方A机器运行着Redis，B机器需要访问，但B ping不通A，A能ping通B

## 使用方式
1. 安装工具
```
go install github.com/byebyebruce/tun/cmd/tun@latest
```
2. B机器运行
`tun s --listen=:9900`
3. A机器运行
`tun c --server=${B_IP}:9900 --remote=:6379 127.0.0.1:6379`
4. B机器连接A机器的redis
`redis-cli -p 6379`

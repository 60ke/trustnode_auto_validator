## 自动添加tm bsc配置

### 编译
`go build -o nodeUpgrade`

### 运行
`./nodeUpgrade`

### 添加节点
```bash
curl --location --request POST '127.0.0.1:6667/add_validators' \
--header 'Content-Type: application/json' \
--data-raw '{"ips":["106.3.133.178","106.3.133.179"],"token":"3D3781351A3EE9E4"}'
```
> ips为待添加验证者节点ip地址列表,"127.0.0.1"为server ip
### 获取当前验证者

```bash
curl 127.0.0.1:6667/get_validators
```
>"127.0.0.1"为server ip


### 日志

日志默认为`update.log`,同时使用`lumberjack`对日志进行切割,单个日志大小限制为5M,时间为24小时
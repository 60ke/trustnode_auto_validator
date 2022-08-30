## 自动添加tm bsc配置

### 编译
`go build -o nodeUpgrade`

### 运行
`./nodeUpgrade`

### 添加节点
- 接口地址: http://127.0.0.1:6667/add_validators
- 请求方式: POST
- 请求参数:
    - ips: 目标ip地址列表
    - token: 服务端token
    - type: 添加ip类型:bsc/tm/all   bsc:bsc节点,tm:tm节点,all:bsc和tm节点
    - action: 更新动作(暂时未使用):add/del  add:添加,del:删除
- 返回结果: 返回结果为json
    - 结果参数:
        - msg: 结果详情
        - status: 结果状态: success/failed
- 添加bsc节点的`curl`示例:
    ```bash
    curl --location --request POST '127.0.0.1:6667/add_validators' \
    --header 'Content-Type: application/json' \
    --data-raw '{"ips":["101.251.207.187"],"token":"3D3781351A3EE9E4","type":"bsc"}'
    ```
> ips为待添加验证者节点ip地址列表,"127.0.0.1"为server ip
### 获取当前验证者

```bash
curl 127.0.0.1:6667/get_validators
```
>"127.0.0.1"为server ip


### 日志

日志默认为`update.log`,同时使用`lumberjack`对日志进行切割,单个日志大小限制为5M,时间为24小时
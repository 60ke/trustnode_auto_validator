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
    - action: 更新动作(del暂时未使用):add/del  add:添加,del:删除
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
- 添加tm节点的`curl`示例:
    ```bash
    curl --location --request POST '127.0.0.1:6667/add_validators' \
    --header 'Content-Type: application/json' \
    --data-raw '{"ips":["101.251.207.187"],"token":"3D3781351A3EE9E4","type":"tm"}'
    ```
- 同时添加tm和bsc节点的`curl`示例:
    ```bash
    curl --location --request POST '127.0.0.1:6667/add_validators' \
    --header 'Content-Type: application/json' \
    --data-raw '{"ips":["101.251.207.187"],"token":"3D3781351A3EE9E4","type":"all"}'
    ```     
> ips为待添加验证者节点ip地址列表,"127.0.0.1"为server ip
### 获取当前验证者

```bash
curl 127.0.0.1:6667/get_validators
```
>"127.0.0.1"为server ip


### 日志

日志默认为`update.log`,同时使用`lumberjack`对日志进行切割,单个日志大小限制为5M,时间为24小时

### 配置
配置文件为当前路径下的`config.toml`文件.`config_ori.toml`为项目的原始(默认)配置文件,有5个验证者.当验证者重置为5个验证者时,`config.toml`文件内容应与`config_ori.toml`一致.
相关配置说明以`config_ori.toml`注释为准


## 其它
对于新增的节点,bsc从创世块同步时,trustconfig需配置为5个,待同步至最新的区块时,trustconfig不会自动更新,当前需要手动修改使其与其它节点一致.
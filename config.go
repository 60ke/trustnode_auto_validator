package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/pelletier/go-toml"
)

var Conf Config

type Host struct {
	Ip    string `toml:"ip" comment:"ip地址"`
	IsNew bool   `toml:"isNew" comment:"用于确定是否为新加ip,默认为false,已弃用"`
}

type TmServerIP struct {
	IP string `toml:"ip"`
}

type ServerConf struct {
	Token string `toml:"token" comment:"bsc链server token"`
}

type Monitor struct {
	DingUrl       string   `toml:"ding-url" comment:"钉钉机器人接口"`
	PrefixKey     string   `toml:"prefix-key" comment:"钉钉安全,消息前缀"`
	Interval      int      `toml:"interval" comment:"获取块高时间间隔,单位为秒"`
	RetryTimes    int      `toml:"retry-times" comment:"判断区块落后所需次数"`
	AbnormalHosts []string `toml:"abnormal-hosts" comment:"落后区块节点ip列表"`
}

type Config struct {
	TM         []Host     `toml:"tm" comment:"tm链节点ip列表"`
	BSC        []Host     `toml:"bsc" comment:"bsc链节点ip列表"`
	Server     ServerConf `toml:"server" comment:"bsc链server配置"`
	TmServer   TmServerIP `toml:"tm-server" comment:"tm链server ip"`
	TmMonitor  Monitor    `toml:"tm-monitor" comment:"tm落后节点监视器"`
	BscMonitor Monitor    `toml:"bsc-monitor" comment:"bsc节点停止出块监视"`
}

func init() {
	LoadConf()
}

func LoadConf() *Config {
	data, _ := os.ReadFile("./config.toml")
	toml.Unmarshal(data, &Conf)

	return &Conf
}

func SaveConf(conf Config) error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	f, err := os.Create("./config.toml")
	if err != nil {
		Logger.Error(err)
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	enc.Order(toml.OrderPreserve)

	if err = enc.Encode(conf); err != nil {
		return err
	}
	return nil
}

// 获取去重后的ip地址列表
func GetIps(nodeType string) ([]string, error) {
	temp := make(map[string]bool)
	var ips []string
	if nodeType == "bsc" {
		for _, host := range Conf.BSC {
			if _, ok := temp[host.Ip]; !ok {
				temp[host.Ip] = host.IsNew
				// 不再支持从配置文件添加
				// if host.IsNew {
				// 	ips = append(ips, host.Ip)
				// }
				ips = append(ips, host.Ip)
			}
		}
	} else if nodeType == "tm" {
		for _, host := range Conf.TM {
			if _, ok := temp[host.Ip]; !ok {
				temp[host.Ip] = host.IsNew
				ips = append(ips, host.Ip)
			}
		}
	} else {
		err := fmt.Errorf("bad nodeType: %s", nodeType)
		Logger.Error(err)
		return nil, err
	}

	return ips, nil
}

func (conf Config) String() string {
	b, _ := toml.Marshal(conf)
	return string(b)
}

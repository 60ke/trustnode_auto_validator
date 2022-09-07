package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type BscNodeStatus struct {
	Height int64
	Ip     string
	// 是否停止出块
	Increase bool
	// 网络等其它错误导致的状态获取异常
	Err error
}

func GetBscHeight(ip string) (int64, error) {
	var res BscResult
	url := "http://" + ip + ":8545"

	payload := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_blockNumber", "id":1}`)

	b, err := post(url, payload)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return 0, err
	}

	res.Result = strings.TrimPrefix(res.Result, "0x")

	if res.Result == "" {
		res.Result = "0"
	}

	h, err := strconv.ParseInt(res.Result, 16, 64)
	if err != nil {
		return 0, err
	}
	Logger.Infof("bscHeight: %s:%d", ip, h)
	return h, nil

}

func BscNodeWatch(ip string, interval, retry int) BscNodeStatus {
	var bscNodeStatus BscNodeStatus
	bscNodeStatus.Increase = false
	bscNodeStatus.Ip = ip
	oldH, err := GetBscHeight(ip)
	if err != nil {
		Logger.Error(err)
		bscNodeStatus.Err = err
	}
	bscNodeStatus.Height = oldH
	times := 1
	if err != nil {
		Logger.Error(err)
	}
	for times < retry {
		newH, _ := GetBscHeight(ip)
		if newH > oldH {
			bscNodeStatus.Increase = true
			return bscNodeStatus
		}
		time.Sleep(time.Duration(interval) * time.Second)
		times += 1
	}
	return bscNodeStatus

}

func BscWatch() {
	for {
		nodeType := "bsc"
		interval := Conf.BscMonitor.Interval
		retry := Conf.BscMonitor.RetryTimes
		prefix := Conf.BscMonitor.PrefixKey
		url := Conf.BscMonitor.DingUrl

		var abnormals []string
		// 读取旧的异常ip列表
		lastAbnormals := Conf.BscMonitor.AbnormalHosts

		var msg DingTmMsg
		msg.MsgType = "text"

		ips, _ := GetIps("bsc")
		Logger.Info("bsc ip list", ips)

		results := make(chan BscNodeStatus, len(ips))
		for _, ip := range ips {
			go func(ip string, interval, retry int) {
				bscNodeWatch := BscNodeWatch(ip, interval, retry)
				results <- bscNodeWatch
			}(ip, interval, retry)
		}

		for i := 0; i < len(ips); i++ {
			result := <-results
			Logger.Debug(result)
			if !result.Increase && result.Err == nil {
				abnormals = append(abnormals, result.Ip)
			}
		}
		Conf.BscMonitor.AbnormalHosts = abnormals
		SaveConf(Conf)
		newAbnormals := getNewAbnormals(lastAbnormals, abnormals)
		normalIp := getNormal(abnormals, "bsc")
		clusterHeight, _ := GetBscHeight(normalIp)
		if len(newAbnormals) != 0 {
			content := GenMsg(newAbnormals, nodeType, prefix, clusterHeight)
			sendMsg(url, prefix, nodeType, content)
		}

	}
}

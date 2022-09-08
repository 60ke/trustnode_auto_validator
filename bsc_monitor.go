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
	Logger.Debugf("bscHeight: %s:%d", ip, h)
	return h, nil

}

func BscNodeWatch(ip string, interval, retry int) BscNodeStatus {
	var bscNodeStatus BscNodeStatus
	bscNodeStatus.Increase = false
	bscNodeStatus.Ip = ip
	start := time.Now()
	oldH, err := GetBscHeight(ip)
	if err != nil {
		Logger.Error(err)
		bscNodeStatus.Err = err
		return bscNodeStatus
	}
	bscNodeStatus.Height = oldH
	times := 0

	for times < retry {
		newH, _ := GetBscHeight(ip)
		if newH > oldH {
			bscNodeStatus.Increase = true
			return bscNodeStatus
		}
		time.Sleep(time.Duration(interval) * time.Second)
		times += 1
	}
	Logger.Infof("bsc节点:%s,%ds内未出新快,块高:%d", ip, int(time.Since(start).Seconds()), oldH)
	return bscNodeStatus

}

func BscWatch() {
	for {
		nodeType := "bsc"
		interval := Conf.BscMonitor.Interval
		retry := Conf.BscMonitor.RetryTimes
		errPrefix := Conf.BscMonitor.ErrPrefixKey
		okPrefix := Conf.BscMonitor.OkPrefixKey
		url := Conf.BscMonitor.DingUrl

		var abnormals []string
		// 读取旧的异常ip列表
		lastAbnormals := Conf.BscMonitor.AbnormalHosts

		var msg DingErrMsg
		msg.MsgType = "text"

		ips, _ := GetIps("bsc")
		Logger.Debug("bsc ip list", ips)

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
		okNodes := getOknodes(lastAbnormals, abnormals)

		// 通知恢复正常的节点
		if len(okNodes) != 0 {
			okContent := GenOkMsg(okNodes, nodeType, okPrefix)
			sendMsg(url, nodeType, okContent)
		}

		// 通知所有节点恢复正常
		if len(abnormals) == 0 {
			if len(lastAbnormals) > 0 {
				okContent := GenOkMsg(nil, nodeType, okPrefix)
				sendMsg(url, nodeType, okContent)
			}
		}

		normalIp := getNormal(abnormals, nodeType)
		clusterHeight, _ := GetBscHeight(normalIp)
		if len(newAbnormals) != 0 {
			content := GenErrMsg(newAbnormals, nodeType, errPrefix, clusterHeight)
			sendMsg(url, nodeType, content)
		}

	}
}

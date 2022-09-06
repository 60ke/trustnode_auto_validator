package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TmH struct {
	Height int64
	Ip     string
}

// tm集群的块高状态
type TmsHStatus struct {
	Hosts []TmH
}

type ChResult struct {
	T   TmH
	Err error
}

type LagNodes map[string]int

type Text struct {
	Content string `json:"content"`
}
type DingMsg struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

func GetTmsHStatus() (TmsHStatus, error) {
	var tms TmsHStatus

	tmIps, _ := GetIps("tm")
	Logger.Info(tmIps)
	results := make(chan ChResult, len(tmIps))
	for _, tmIp := range tmIps {
		go func(ip string) {
			var res TmHResponse

			var chResult ChResult
			chResult.T.Ip = ip
			url := "http://" + ip + ":46657/tri_abci_info"
			r, err := get(url)
			if err != nil {
				chResult.Err = err
				results <- chResult
				return
			}
			err = json.Unmarshal(r, &res)
			if err != nil {
				Logger.Error(err)
				chResult.Err = err
				results <- chResult
				return
			}
			if res.Result.Response.LastBlockHeight == "" {
				res.Result.Response.LastBlockHeight = "0"
			}
			h, err := strconv.ParseInt(res.Result.Response.LastBlockHeight, 10, 64)
			if err != nil {
				Logger.Error(err)
				chResult.Err = err
				results <- chResult
				return
			}
			chResult.T.Height = h
			Logger.Infof("tmHeight: %s:%d", chResult.T.Ip, chResult.T.Height)
			results <- chResult
		}(tmIp)
	}

	for i := 0; i < len(tmIps); i++ {
		result := <-results
		Logger.Debug(result)
		if result.Err == nil {
			tms.Hosts = append(tms.Hosts, result.T)
		}
	}
	close(results)
	return tms, nil
}

func getMaxH(tms TmsHStatus) int64 {
	var max int64
	for _, host := range tms.Hosts {
		if host.Height > max {
			max = host.Height
		}
	}
	return max
}

func getLagNodes(lagNodes LagNodes) LagNodes {
	tms, err := GetTmsHStatus()
	if err != nil {
		Logger.Error(err)
	}
	Logger.Info(tms)
	maxH := getMaxH(tms)
	for _, host := range tms.Hosts {
		if host.Height < maxH {
			lagNodes[host.Ip] += 1
		}
	}
	Logger.Info(lagNodes)
	return lagNodes
}

func getNewAbnormals(lastAbnormals, abnormals []string) []string {
	var newAbnormals []string
	for _, ip := range abnormals {
		var in bool
		for _, lastIp := range lastAbnormals {
			if lastIp == ip {
				in = true
			}
		}
		if !in {
			newAbnormals = append(newAbnormals, ip)
		}
	}
	return newAbnormals
}

// 发送钉钉通知
func sendMsg(url, prefix, content string) {
	Logger.Info("sendMsg:", content)
	payload := strings.NewReader(fmt.Sprintf(`{"msgtype": "text","text": {"content":"%s"}}`, prefix+content))
	post(url, payload)
}

// 获取落后节点
func GetLagNodes() {
	// 先从配置中读取
	for {
		times := 0

		var abnormals []string
		// 读取旧的异常ip列表
		lastAbnormals := Conf.Monitor.AbnormalHosts
		interval := Conf.Monitor.Interval
		retry := Conf.Monitor.RetryTimes
		prefix := Conf.Monitor.PrefixKey
		url := Conf.Monitor.DingUrl
		var lagNodes = make(LagNodes)
		Logger.Info(interval, retry, prefix, url)
		for times < retry {
			lagNodes = getLagNodes(lagNodes)
			time.Sleep(time.Duration(interval) * time.Second)
			times += 1
		}
		for ip, timesInt := range lagNodes {
			Logger.Info(ip, timesInt)
			if timesInt == retry {
				abnormals = append(abnormals, ip)
			}
		}
		Conf.Monitor.AbnormalHosts = abnormals
		SaveConf(Conf)
		newAbnormals := getNewAbnormals(lastAbnormals, abnormals)
		if len(abnormals) != 0 {
			sendMsg(url, prefix, strings.Join(newAbnormals, ","))
		}

		times = 0

	}
	// return lasts, nil
}

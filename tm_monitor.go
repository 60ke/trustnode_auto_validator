package main

import (
	"encoding/json"
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

type TmErrHost struct {
	Title         string
	IP            string
	LocalHeight   int64
	ClusterHeight int64
}
type Text struct {
	Content []TmErrHost `json:"content"`
}
type DingTmMsg struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

func GetTmHeight(ip string) (int64, error) {

	var res TmHResponse
	url := "http://" + ip + ":46657/tri_abci_info"
	r, err := get(url)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(r, &res)
	if err != nil {
		return 0, err
	}
	if res.Result.Response.LastBlockHeight == "" {
		res.Result.Response.LastBlockHeight = "0"
	}
	h, err := strconv.ParseInt(res.Result.Response.LastBlockHeight, 10, 64)
	if err != nil {
		return 0, err
	}
	return h, nil

}

func GetTmsHStatus() (TmsHStatus, error) {
	var tms TmsHStatus

	tmIps, _ := GetIps("tm")
	Logger.Info(tmIps)
	results := make(chan ChResult, len(tmIps))
	for _, tmIp := range tmIps {
		go func(ip string) {
			var chResult ChResult
			chResult.T.Ip = ip

			h, err := GetTmHeight(ip)
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
func sendTmMsg(url, prefix, content string) {
	Logger.Info("sendTmMsg:", content)
	payload := strings.NewReader(content)
	post(url, payload)
}

// 获取正常的节点ip
func getNormal(ips []string) string {
	tmIps, _ := GetIps("tm")
	for _, tmIp := range tmIps {
		in := false
		for _, ip := range ips {
			if tmIp == ip {
				in = true
			}
		}
		if !in {
			return tmIp
		}
	}
	Logger.Error("cant get abnormal ip")
	return ""
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
		normalIp := getNormal(abnormals)
		clusterHeight, _ := GetTmHeight(normalIp)
		if len(newAbnormals) != 0 {
			var msg DingTmMsg
			msg.MsgType = "text"
			for _, ip := range newAbnormals {
				var host TmErrHost
				localHeight, _ := GetTmHeight(ip)
				host.ClusterHeight = clusterHeight
				host.LocalHeight = localHeight
				host.IP = ip
				host.Title = prefix
				msg.Text.Content = append(msg.Text.Content, host)
			}
			content, _ := json.Marshal(msg)
			sendTmMsg(url, prefix, string(content))
		}

		times = 0

	}
	// return lasts, nil
}

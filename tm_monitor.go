package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// 节点块高
type NodeH struct {
	Height int64
	Ip     string
}

// 集群的块高状态
type ClusterHStatus struct {
	Nodes []NodeH
}

// 获取节点块高,channel返回结果
type ChResult struct {
	H   NodeH
	Err error
}

// 落后节点map
type LagNodes map[string]int

// 高度错误节点
type HeightErrHost struct {
	Title         string
	IP            string
	LocalHeight   int64
	ClusterHeight int64
}

// 钉钉通知text
type Text struct {
	Content []HeightErrHost `json:"content"`
}

// 钉钉通知消息
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

func GetClusterHStatus(nodeType string) (ClusterHStatus, error) {
	var cluster ClusterHStatus

	ips, _ := GetIps(nodeType)
	Logger.Info("tm ip list :", ips)
	results := make(chan ChResult, len(ips))
	for _, tmIp := range ips {
		go func(ip string) {
			var chResult ChResult
			chResult.H.Ip = ip

			h, err := GetTmHeight(ip)
			if err != nil {
				Logger.Error(err)
				chResult.Err = err
				results <- chResult
				return
			}
			chResult.H.Height = h
			Logger.Infof("tmHeight: %s:%d", chResult.H.Ip, chResult.H.Height)
			results <- chResult
		}(tmIp)
	}

	for i := 0; i < len(ips); i++ {
		result := <-results
		Logger.Debug(result)
		if result.Err == nil {
			cluster.Nodes = append(cluster.Nodes, result.H)
		}
	}
	close(results)
	return cluster, nil
}

func getMaxH(tms ClusterHStatus) int64 {
	var max int64
	for _, host := range tms.Nodes {
		if host.Height > max {
			max = host.Height
		}
	}
	return max
}

// 获取落后节点
func getLagNodes(lagNodes LagNodes, nodeType string) LagNodes {
	tms, err := GetClusterHStatus(nodeType)
	if err != nil {
		Logger.Error(err)
	}
	Logger.Info(tms)
	maxH := getMaxH(tms)
	for _, host := range tms.Nodes {
		if host.Height < maxH {
			lagNodes[host.Ip] += 1
		}
	}
	Logger.Info(lagNodes)
	return lagNodes
}

// 获取新的异常节点列表
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

func GetHeight(nodeType, ip string) (int64, error) {
	if nodeType == "tm" {
		return GetTmHeight(ip)
	}
	return GetBscHeight(ip)
}

func GenMsg(errNodes []string, nodeType, msgPrefix string, clusterHeight int64) string {
	var msg DingTmMsg
	msg.MsgType = "text"
	for _, ip := range errNodes {
		var host HeightErrHost
		localHeight, _ := GetHeight(nodeType, ip)
		host.ClusterHeight = clusterHeight
		host.LocalHeight = localHeight
		host.IP = ip
		host.Title = msgPrefix
		msg.Text.Content = append(msg.Text.Content, host)
	}
	content, _ := json.Marshal(msg)
	return string(content)
}

// 发送钉钉通知
func sendMsg(url, prefix, nodeType, content string) {
	Logger.Info("send", nodeType, "Msg:", content)
	payload := strings.NewReader(content)
	post(url, payload)
}

// 获取正常的节点ip
func getNormal(ips []string, nodeType string) string {
	cluster, _ := GetIps(nodeType)
	for _, node := range cluster {
		in := false
		for _, ip := range ips {
			if node == ip {
				in = true
			}
		}
		if !in {
			return node
		}
	}
	Logger.Error("cant get abnormal ip")
	return ""
}

// 获取落后节点
func TMWatch() {
	// 先从配置中读取
	for {
		times := 0
		nodeType := "tm"
		var abnormals []string
		// 读取旧的异常ip列表
		lastAbnormals := Conf.TmMonitor.AbnormalHosts
		interval := Conf.TmMonitor.Interval
		retry := Conf.TmMonitor.RetryTimes
		prefix := Conf.TmMonitor.PrefixKey
		url := Conf.TmMonitor.DingUrl
		var lagNodes = make(LagNodes)
		Logger.Info(interval, retry, prefix, url)
		for times < retry {
			lagNodes = getLagNodes(lagNodes, nodeType)
			time.Sleep(time.Duration(interval) * time.Second)
			times += 1
		}
		for ip, timesInt := range lagNodes {
			Logger.Info(ip, timesInt)
			if timesInt == retry {
				abnormals = append(abnormals, ip)
			}
		}
		Conf.TmMonitor.AbnormalHosts = abnormals
		SaveConf(Conf)
		newAbnormals := getNewAbnormals(lastAbnormals, abnormals)
		normalIp := getNormal(abnormals, nodeType)
		clusterHeight, _ := GetTmHeight(normalIp)
		if len(newAbnormals) != 0 {
			content := GenMsg(newAbnormals, nodeType, prefix, clusterHeight)
			sendMsg(url, prefix, nodeType, content)
		}

		times = 0

	}
	// return lasts, nil
}

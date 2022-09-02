package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var Version string
var Succ string = "success"
var Fail string = "failed"

type IPData struct {
	IPs    []string `json:"ips"`
	Type   string   `json:"type"` //tm or bsc or all
	Token  string   `json:"token"`
	Action string   `json:"action"` //add or del
}

func addBsc(newIp string) error {
	ips, _ := GetIps("bsc")

	// 拒绝请求ip已在当前验证者列表中
	if CheckIn(newIp, ips) {
		err := fmt.Errorf("addBsc err: %s is already exsit", newIp)
		Logger.Error(err)
		return err
	}

	newConf := Conf
	host := Host{
		Ip:    newIp,
		IsNew: false,
	}
	newConf.BSC = append(newConf.BSC, host)
	ips = append(ips, newIp)

	bscTmMap, err := GenMap([]string{newIp})
	if err != nil {
		return err
	}

	// 向bsc发送节点添加请求,重试次数为5次
	failure := PostNode(ips, bscTmMap, 5)
	if len(failure) != 0 {
		err := fmt.Errorf("addBsc err: ips: %s", strings.Join(failure, ","))
		Logger.Error(err)
		return err
	}

	// 保存bsc到config
	Conf = newConf
	SaveConf(Conf)
	return nil
}

func addTm(newIp string) error {
	ips, _ := GetIps("tm")

	// 拒绝请求ip已在当前验证者列表中
	if CheckIn(newIp, ips) {
		err := fmt.Errorf("addTm err: %s is already exsit", newIp)
		Logger.Error(err)
		return err
	}
	// Logger.Info(Conf)
	newConf := Conf
	host := Host{
		Ip:    newIp,
		IsNew: false,
	}
	newConf.TM = append(newConf.TM, host)
	var tmErrResult TmTxErrResult
	ret, err := SendTmTx(newConf.TmServer.IP, newIp)
	if err != nil {
		return err
	}

	// 如果序列化成功表明tx交易有错误返回
	tmErr := json.Unmarshal(ret, &tmErrResult)
	Logger.Info(tmErr, tmErrResult.Error)
	if tmErrResult.Error.Code != 0 {
		return fmt.Errorf(string(ret))
	}
	Conf = newConf
	err = SaveConf(Conf)
	if err != nil {
		return err
	}
	return nil

}

// !!!当前的处理逻辑是bsc,tm完全一一对应,故为了保证原子性,节点添加需逐个添加.
func AddValidators(c *gin.Context) {
	var ipdata IPData
	var bscfails []string
	var tmfails []string
	var err error

	var addResult AddResult
	addResult.Status = Succ
	if err := c.BindJSON(&ipdata); err != nil {
		return
	}
	if ipdata.Token != Conf.Server.Token {
		Logger.Debug(Conf.Server.Token)
		addResult.Status = Fail
		addResult.Msg = "error token:" + ipdata.Token
		c.IndentedJSON(http.StatusBadRequest, addResult)
		return
	}

	switch ipdata.Type {
	case "bsc":
		for _, ip := range ipdata.IPs {
			var addHostResult AddHostResult
			addHostResult.IP = ip
			err = addBsc(ip)
			addHostResult.AddBsc = Succ
			if err != nil {
				Logger.Error("addBsc err:", err)
				addHostResult.AddBsc = Fail
				addHostResult.BscErr = err.Error()
				addResult.Hosts = append(addResult.Hosts, addHostResult)
				bscfails = append(bscfails, ip)

			}
			addResult.Hosts = append(addResult.Hosts, addHostResult)
		}
		if len(bscfails) != 0 {
			addResult.Status = Fail
			addResult.Msg = "some addbsc failed"
			c.IndentedJSON(http.StatusOK, addResult)
			return
		}
	case "tm":
		for _, ip := range ipdata.IPs {
			var addHostResult AddHostResult
			addHostResult.IP = ip
			err = addTm(ip)
			addHostResult.AddTm = Succ
			if err != nil {
				Logger.Error("addTm err:", err)
				addHostResult.AddTm = Fail
				addHostResult.TmErr = err.Error()
				tmfails = append(tmfails, ip)
			}
			addResult.Hosts = append(addResult.Hosts, addHostResult)
		}
		if len(tmfails) != 0 {
			addResult.Status = Fail
			addResult.Msg = "some ip addtm failed"
			c.IndentedJSON(http.StatusOK, addResult)
			return
		}
	case "all":
		for _, ip := range ipdata.IPs {
			var addHostResult AddHostResult
			addHostResult.IP = ip
			addHostResult.AddBsc = Succ
			addHostResult.AddTm = Succ

			tmErr := addTm(ip)
			if tmErr != nil {
				addHostResult.AddTm = Fail
				addHostResult.TmErr = tmErr.Error()
				tmfails = append(tmfails, ip)
				addHostResult.AddBsc = Fail
				addHostResult.BscErr = "addtm failed"
				bscfails = append(bscfails, ip)
			} else {
				bscErr := addBsc(ip)
				if bscErr != nil {
					addHostResult.AddBsc = Fail
					addHostResult.BscErr = bscErr.Error()
					bscfails = append(bscfails, ip)
				}
			}
			addResult.Hosts = append(addResult.Hosts, addHostResult)
		}

		if len(bscfails) != 0 {
			addResult.Status = Fail
			if len(tmfails) == 0 {
				addResult.Msg = "some addbsc failed"
				c.IndentedJSON(http.StatusOK, addResult)
				return
			} else {
				addResult.Msg = "some addbsc failed,some addtm failed"
				c.IndentedJSON(http.StatusOK, addResult)
				return
			}

		} else {
			if len(tmfails) != 0 {
				addResult.Status = Fail
				addResult.Msg = "some addtm failed"
				c.IndentedJSON(http.StatusOK, addResult)
				return
			}
		}
	default:
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": Fail, "msg": "error type:" + ipdata.Type})
		return
	}

	c.IndentedJSON(http.StatusOK, addResult)

}

func GetValidators(c *gin.Context) {
	bscIps, _ := GetIps("bsc")
	tmIps, _ := GetIps("tm")
	c.IndentedJSON(http.StatusOK, gin.H{"bsc": bscIps, "tm": tmIps})
}

func Server() {
	fmt.Printf("Version: %s\n", Version)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/add_validators", AddValidators)
	r.GET("/get_validators", GetValidators)
	r.Run(":6667") // listen for incoming connections
}

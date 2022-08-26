package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

	newConf := Conf
	host := Host{
		Ip:    newIp,
		IsNew: false,
	}
	newConf.TM = append(newConf.TM, host)
	_, err := SendTmTx(ips[0], newIp)
	if err != nil {
		return err
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
	if err := c.BindJSON(&ipdata); err != nil {
		return
	}
	if ipdata.Token != Conf.Server.Token {
		Logger.Debug(Conf.Server.Token)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "error token:" + ipdata.Token})
		return
	}
	var bscfails []string
	var tmfails []string
	switch ipdata.Type {
	case "bsc":
		for _, ip := range ipdata.IPs {
			err := addBsc(ip)
			if err != nil {
				bscfails = append(bscfails, ip)
			}
		}
		if len(bscfails) != 0 {
			c.IndentedJSON(http.StatusOK, gin.H{"status": "failed", "msg": "some ip add failed: " + strings.Join(bscfails, ",")})
			return
		}
	case "tm":
		for _, ip := range ipdata.IPs {
			err := addTm(ip)
			if err != nil {
				tmfails = append(tmfails, ip)
			}
		}
		if len(tmfails) != 0 {
			c.IndentedJSON(http.StatusOK, gin.H{"status": "failed", "msg": "some ip add failed: " + strings.Join(tmfails, ",")})
			return
		}
	case "all":
		for _, ip := range ipdata.IPs {
			err := addBsc(ip)
			if err != nil {
				bscfails = append(bscfails, ip)
			}
		}

		for _, ip := range ipdata.IPs {
			err := addTm(ip)
			if err != nil {
				tmfails = append(tmfails, ip)
			}
		}
		if len(bscfails) != 0 {
			if len(tmfails) == 0 {
				c.IndentedJSON(http.StatusOK, gin.H{"status": "failed", "msg": "some ip addbsc failed: " + strings.Join(bscfails, ",")})
				return
			} else {
				c.IndentedJSON(http.StatusOK, gin.H{"status": "failed", "msg": "some ip addbsc failed: " + strings.Join(bscfails, ",") + "some ip addtm failed: " + strings.Join(tmfails, ",")})
				return
			}

		} else {
			if len(tmfails) != 0 {
				c.IndentedJSON(http.StatusOK, gin.H{"status": "failed", "msg": "some ip addtm failed: " + strings.Join(tmfails, ",")})
				return
			}
		}
	default:
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "error type:" + ipdata.Type})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "msg": ipdata.IPs})

}

func GetValidators(c *gin.Context) {
	ips, _ := GetIps("bsc")
	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "ips": ips})
}

func Server() {
	r := gin.Default()
	r.POST("/add_validators", AddValidators)
	r.GET("/get_validators", GetValidators)
	r.Run(":6667") // listen for incoming connections
}

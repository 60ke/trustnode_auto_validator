package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func get(url string) ([]byte, error) {
	// url := "http://106.3.133.179:46657/tri_block_info?height=104360"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func post(url string, payload *strings.Reader) ([]byte, error) {

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetTmAddr(ip string) (string, error) {
	url := "http://" + ip + ":46657/tri_status?"

	var ret TmResult
	b, err := get(url)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return "", err
	}
	return ret.Result.ValidatorInfo.Address, nil
}

func GetTmPK(ip string) (string, error) {
	url := "http://" + ip + ":46657/tri_pubkey?"
	var ret TmPub
	b, err := get(url)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return "", err
	}
	pubStr, err := base64.StdEncoding.DecodeString(ret.Result.PubKey.Value)
	if err != nil {
		return "", err
	}
	data := "192.168.1.227/" + fmt.Sprintf("%X", pubStr) + "/100"
	return data, nil
}

func SendTmTx(ip string, node string) ([]byte, error) {
	data, err := GetTmPK(node)
	if err != nil {
		return nil, err
	}
	url := "http://" + ip + "tri_broadcast_tx_commit/?tx=\"[addvalidator]val:" + data + "\""
	Logger.Debug(url)
	return nil, nil
	// return get(url)
	// ret, err := get(url)
	// Logger.Info(string(ret))
	// if err != nil {
	// 	return nil, err
	// }
}

func GetBscAddr(ip string) (string, error) {
	url := "http://" + ip + ":8545"
	payload := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_coinbase", "id":1}`)
	var ret BscResult
	b, err := post(url, payload)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return "", err
	}
	return ret.Result, nil
}

// 生成bsc tm地址对
func GenMap(ips []string) (BscTmMap, error) {
	Logger.Info("GenMap:", ips)
	var bscTmMap BscTmMap
	for _, ip := range ips {
		var btm BTM
		addr := net.ParseIP(ip)
		if addr == nil {
			return BscTmMap{}, fmt.Errorf("invalid ip: %s", ip)
		}
		tmAddr, err := GetTmAddr(ip)
		if err != nil {
			return BscTmMap{}, err
		}
		bscAddr, err := GetBscAddr(ip)
		if err != nil {
			return BscTmMap{}, err
		}
		btm.TmNodePubkeyAddr = tmAddr
		btm.BscNodePubkeyAddr = bscAddr
		bscTmMap.BscPubkeyAddrMaps = append(bscTmMap.BscPubkeyAddrMaps, btm)

	}
	b, err := json.Marshal(bscTmMap)
	if err != nil {
		return BscTmMap{}, err
	}
	Logger.Info("newbscTmMap:", string(b))
	return bscTmMap, nil
}

// 添加验证者
func PostNode(ips []string, bscTmMap BscTmMap, retrytimes int) []string {
	var toPost []string
	toPost = ips
	var retry int
	for len(toPost) != 0 || retry < retrytimes {
		var failures []string
		for _, ip := range toPost {
			url := "http://" + ip + ":6666" + "/upgrade/addvalidatorv2"
			num := len(bscTmMap.BscPubkeyAddrMaps)
			json_str, _ := json.Marshal(bscTmMap)
			payload := strings.NewReader(fmt.Sprintf(`{"AccessToken":%s,"bscTMPubkeyPairs": %s, "pubkeyNum": %d}`, Conf.Server.Token, string(json_str), num))
			_, err := post(url, payload)
			if err != nil {
				failures = append(failures, ip)
			}
		}
		toPost = failures
		retry += 1
	}

	return toPost
}

// 判断string元素是否在列表中
func CheckIn(e string, l []string) bool {
	for _, v := range l {
		if v == e {
			return true
		}
	}
	return false
}

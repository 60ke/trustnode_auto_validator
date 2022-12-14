package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func get(url string) ([]byte, error) {
	// url := "http://106.3.133.179:46657/tri_block_info?height=104360"

	client := &http.Client{}
	client.Timeout = time.Second * 5
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

	var ret TmSResponse
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
	data := ip + "/" + fmt.Sprintf("%X", pubStr) + "/100"
	return data, nil
}

func SendTmTx(ip string, node string) ([]byte, error) {
	data, err := GetTmPK(node)
	if err != nil {
		return nil, err
	}
	url := "http://" + ip + ":46657/tri_broadcast_tx_commit?tx=\"[addvalidator]val:" + data + "\""
	Logger.Debug("send TmTx:", url)

	return get(url)
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

// ??????bsc tm?????????
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
			Logger.Error(err)
			return BscTmMap{}, err
		}
		bscAddr, err := GetBscAddr(ip)
		if err != nil {
			Logger.Error(err)
			return BscTmMap{}, err
		}
		tmAddr = strings.ToLower(tmAddr)
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

// ???????????????
func PostNode(ips []string, bscTmMap BscTmMap, retrytimes int) []string {
	Logger.Info("start add bsc to:", strings.Join(ips, ","))
	var toPost []string
	toPost = ips
	var retry int
	for len(toPost) != 0 && retry < retrytimes {
		var failures []string
		for _, ip := range toPost {
			url := "http://" + ip + ":6666" + "/upgrade/addvalidatorv2"
			num := len(bscTmMap.BscPubkeyAddrMaps)
			json_str, _ := json.Marshal(bscTmMap.BscPubkeyAddrMaps)
			Logger.Info(string(json_str))
			payload := strings.NewReader(fmt.Sprintf(`{"AccessToken":"%s","bscTMPubkeyPairs": %s, "pubkeyNum": %d}`, Conf.Server.Token, string(json_str), num))
			ret, err := post(url, payload)
			if err != nil {
				Logger.Error("add bsc failed,ip : ", ip, ",retry.")
				failures = append(failures, ip)
			}
			if !strings.Contains(string(ret), "????????????") {
				failures = append(failures, ip)
			}
			Logger.Info("add bsc to ", ip, " succ:", string(ret))
		}
		toPost = failures
		retry += 1
	}

	return toPost
}

// ??????string????????????????????????
func CheckIn(e string, l []string) bool {
	for _, v := range l {
		if v == e {
			return true
		}
	}
	return false
}

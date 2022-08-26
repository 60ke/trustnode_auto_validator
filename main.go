package main

// 匹配新增的增量更新

func main() {

	// LoadConf()
	// ips := GetIps()
	// fmt.Println(ips)
	// GenMap(ips)
	go watch()
	Server()
}

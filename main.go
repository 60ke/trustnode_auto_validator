package main

// 匹配新增的增量更新

func main() {

	// go ConfigWatch()
	go TMWatch()
	go BscWatch()
	NodeUpdateServer()
}

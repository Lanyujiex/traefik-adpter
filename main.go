package main

import (
	"log"
	"traefik-adpter/services/listwatch"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("Init")
	stopCh := make(chan struct{})
	defer close(stopCh)
	listwatch.ListIngress(stopCh)

	// 等待程序退出信号
	select {}
}

package main

import (
	"H3C_DDNS/api"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

// http://127.0.0.1:10000/update?server=服务商&domain=<h>&ip=<a>
// 服务商:dnspod 附加参数:secretId
//       alidns

func handler(_ http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	request := r.URL.Query()

	server := request.Get("server")
	domain := request.Get("domain")

	ip := request.Get("ip")

	BaseAuth, _ := base64.StdEncoding.DecodeString(r.Header.Get("Authorization")[6:])
	auth := strings.Split(string(BaseAuth), ":")

	switch server {
	case "dnspod":
		api.DnspodUpdate(r.URL.Query().Get("secretId"), auth[1], domain, ip)
		return
	case "alidns":
		api.AlidnsUpdate(auth[0], auth[1], domain, ip)

		return
	}
}

func main() {
	// 设置路由和处理函数
	http.HandleFunc("/update", handler)

	// 启动服务器
	log.Println("Starting server on :10000")
	err := http.ListenAndServe(":10000", nil)
	if err != nil {
		log.Println("Error starting server:", err)
	}
}

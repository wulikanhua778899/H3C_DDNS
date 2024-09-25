package util

import (
	"log"
	"regexp"
)

// 分割域名
func SplitDomain(domain string) (string, string) {
	// 正则表达式匹配二级域名和主域名
	re := regexp.MustCompile(`^(?P<subdomain>.+)\.(?P<domain>[^.]+\.[^.]+)$`)

	// 使用正则表达式进行匹配
	match := re.FindStringSubmatch(domain)
	if match != nil {
		// 提取二级域名和子域名
		subDomain := match[1]
		mainDomain := match[2]

		return subDomain, mainDomain
	}
	log.Printf("未能匹配域名: %s\n", domain)
	return "", ""
}

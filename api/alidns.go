package api

import (
	"H3C_DDNS/util"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	teaUtil "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"log"
	"strings"
)

func alidnsCreateClient(accessKeyId string, accessKeySecret string) (_result *alidns20150109.Client) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: tea.String(accessKeyId),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: tea.String(accessKeySecret),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Alidns
	config.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	_result = &alidns20150109.Client{}

	_result, _ = alidns20150109.NewClient(config)
	return _result
}

func AlidnsUpdate(accessKeyId string, accessKeySecret string, domain string, m_ip string) {
	client := alidnsCreateClient(accessKeyId, accessKeySecret)

	subDomain, mainDomain := util.SplitDomain(domain)
	if strings.Compare(mainDomain, "") == 0 {
		return
	}

	ip, recordId := alidnsGetIP(client, mainDomain, subDomain)
	if strings.Compare(recordId, "") == 0 {
		return
	}

	if strings.Compare(ip, m_ip) != 0 {
		alidnsSetIP(client, subDomain, m_ip, recordId)
		log.Printf("已更新域名: %s, 原IP: %s, 现IP: %s\n", domain, ip, m_ip)
	} else {
		log.Printf("无需更新域名: %s, IP: %s\n", domain, ip)
	}
}

func alidnsSetIP(client *alidns20150109.Client, subDomain string, ip string, recordID string) {
	updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: tea.String(recordID),
		RR:       tea.String(subDomain),
		Type:     tea.String("A"),
		Value:    tea.String(ip),
	}
	runtime := &teaUtil.RuntimeOptions{}
	_, _err := client.UpdateDomainRecordWithOptions(updateDomainRecordRequest, runtime)
	if _err != nil {
		log.Printf("An API error has returned: %s\n", _err)
	}
}

func alidnsGetIP(client *alidns20150109.Client, mainDomain string, subDomain string) (string, string) {
	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{
		DomainName: tea.String(mainDomain),
		PageSize:   tea.Int64(500),
	}
	runtime := &teaUtil.RuntimeOptions{}

	resp, _err := client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
	if _err != nil {
		log.Printf("An API error has returned: %s\n", _err)
	}

	for _, record := range resp.Body.DomainRecords.Record {
		if *record.RR == subDomain && *record.Status == "ENABLE" && *record.Type == "A" {
			return *record.Value, *record.RecordId
		}
	}
	log.Printf("在 %s 上不存在子域名 %s\n", mainDomain, subDomain)
	return "", ""
}

package api

import (
	"H3C_DDNS/util"
	"encoding/json"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"log"
	"strings"
)

// 定义结构体以映射 JSON 数据
type Response struct {
	RecordCountInfo RecordCountInfo `json:"RecordCountInfo"`
	RecordList      []Record        `json:"RecordList"`
	RequestId       string          `json:"RequestId"`
}

type RecordCountInfo struct {
	SubdomainCount int `json:"SubdomainCount"`
	ListCount      int `json:"ListCount"`
	TotalCount     int `json:"TotalCount"`
}

type Record struct {
	RecordId      int    `json:"RecordId"`
	Value         string `json:"Value"`
	Status        string `json:"Status"`
	UpdatedOn     string `json:"UpdatedOn"`
	Name          string `json:"Name"`
	Line          string `json:"Line"`
	LineId        string `json:"LineId"`
	Type          string `json:"Type"`
	MonitorStatus string `json:"MonitorStatus"`
	Remark        string `json:"Remark"`
	TTL           int    `json:"TTL"`
	MX            int    `json:"MX"`
	DefaultNS     bool   `json:"DefaultNS"`
}

type ApiResponse struct {
	Response Response `json:"Response"`
}

func dnspodCreateClient(secretId string, secretKey string) *dnspod.Client {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)
	return client
}

func DnspodUpdate(secretId string, secretKey string, domain string, m_ip string) {
	client := dnspodCreateClient(secretId, secretKey)

	subDomain, mainDomain := util.SplitDomain(domain)
	if strings.Compare(mainDomain, "") == 0 {
		return
	}

	ip, recordId := dnspodGetIP(client, mainDomain, subDomain)
	if recordId == 0 {
		return
	}

	if strings.Compare(ip, m_ip) != 0 {
		dnspodSetIP(client, mainDomain, subDomain, m_ip, recordId)
		log.Printf("已更新域名: %s, 原IP: %s, 现IP: %s\n", domain, ip, m_ip)
	} else {
		log.Printf("无需更新域名: %s, IP: %s\n", domain, ip)
	}
}

func dnspodSetIP(client *dnspod.Client, mainDomain string, subDomain string, ip string, recordID uint64) string {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyDynamicDNSRequest()

	request.Domain = common.StringPtr(mainDomain)
	request.SubDomain = common.StringPtr(subDomain)
	request.RecordId = common.Uint64Ptr(recordID)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)

	// 返回的resp是一个ModifyDynamicDNSResponse的实例，与请求对象对应
	_, err := client.ModifyDynamicDNS(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s\n", err)
		return ""
	}
	if err != nil {
		panic(err)
	}
	return ""
}

func dnspodGetIP(client *dnspod.Client, mainDomain string, subDomain string) (string, uint64) {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewDescribeRecordListRequest()

	request.Domain = common.StringPtr(mainDomain)

	// 返回的resp是一个DescribeRecordListResponse的实例，与请求对象对应
	response, err := client.DescribeRecordList(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s\n", err)
		return "", 0
	}
	if err != nil {
		panic(err)
	}

	// 创建 ApiResponse 实例
	var apiResponse ApiResponse

	// 解析 JSON 数据
	err = json.Unmarshal([]byte(response.ToJsonString()), &apiResponse)
	if err != nil {
		log.Printf("JSON 解析失败: %s\n", err)
	}

	for _, record := range apiResponse.Response.RecordList {
		if record.Name == subDomain && record.Status == "ENABLE" && record.Type == "A" {
			return record.Value, uint64(record.RecordId)
		}
	}
	log.Printf("在 %s 上不存在子域名 %s\n", mainDomain, subDomain)
	return "", 0
}

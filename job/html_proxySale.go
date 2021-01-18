package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"strings"
)

type proxySale struct {
	Spider
}

func (s *proxySale) StartUrl() []string {
	return []string{
		"https://free.proxy-sale.com/http/?proxy_country=%5B%22%22%5D&proxy_type=%5B%221%22%2C%222%22%5D",
		"https://free.proxy-sale.com/http/?proxy_country=%5B%22%22%5D&proxy_type=%5B%221%22%2C%222%22%5D&proxy_page=2",
	}
}

func (s *proxySale) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *proxySale) Cron() string {
	return "@every 10m"
}

func (s *proxySale) GetReferer() string {
	return "https://free.proxy-sale.com"
}

func (s *proxySale) Run() {
	getProxy(s)
}

func (s *proxySale) Name() string {
	return "proxySale"
}

func (s *proxySale) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table/tbody/tr")
	for _, n := range list {
		portTmp := htmlquery.FindOne(n, "//td[2]//img/@src")

		if portTmp == nil {
			// 解析代理字符串失败
			continue
		}

		// 此站的端口是图片
		portImg := htmlquery.InnerText(portTmp)

		var port = ""
		for key, value := range proxySaleCode {
			if strings.Contains(portImg, key) {
				// 通过图片名称来判断端口
				port = value
				break
			}
		}

		if port == "" {
			continue
		}

		ip := util.FindIp(htmlquery.InnerText(n))

		if ip == "" {
			continue
		}

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

var proxySaleCode = map[string]string{
	"3f4bca399a1445c7d8f231300e51ec44": "9999",
	"4f01e9e5860bda7ff7dd578ccbe6974a": "3000",
	"24fb29b5aeffaffa80d7261712723246": "82",
	"59f3846408407d3143db9f405531ec65": "8888",
	"89b09fc0f66997e8421d8061030c983a": "8080",
	"299db15b3a5f119636dba89962650a03": "8123",
	"503f4e747c3a149fbb5aba851aebefea": "4216",
	"924fee2fff4c5d149433de816d985a8b": "80",
	"24876a6e37e73dfc60446e6d2d27f511": "8081",
	"a089dd618ecb532de560880729150af9": "8118",
	"aae536e39151eacc38068a172533e5fe": "8090",
	"b8c08b828f5c8138a86e18667aa60824": "808",
	"b0450a74d5c38f7859f8066b0a81c4f5": "8088",
	"b806c2a582f29f26d02c4b1d2e470298": "8908",
	"d35e553b965c42349f5740114f06b586": "8089",
	"e3f65f8ff04504869f3e163d532381f2": "3128",
	"e58a3cb12d4b567523fec022a24c254a": "1080",
	"fb1c35f2c2c59bd8d1fdac6899f4ce2b": "8060",
}

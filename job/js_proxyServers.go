package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type proxyServers struct {
	Spider
}

func (s *proxyServers) StartUrl() []string {
	return []string{
		"https://proxyservers.pro/proxy/list/updated/900/protocol/http%2Chttps/order/updated/order_dir/desc/page/1",
	}
}

func (s *proxyServers) Profile() *Setting {
	return &Setting{
		Protocol:    "RenderFetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *proxyServers) Cron() string {
	return "@every 3m"
}

func (s *proxyServers) GetReferer() string {
	return "https://proxyservers.pro"
}

func (s *proxyServers) Run() {
	getProxy(s)
}

func (s *proxyServers) Name() string {
	return "proxyServers"
}

func (s *proxyServers) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='table table-hover']/tbody/tr")
	for _, n := range list {
		ipTmp := htmlquery.FindOne(n, "//td[2]")
		portTmp := htmlquery.FindOne(n, "//td[3]")

		if ipTmp == nil || portTmp == nil {
			// 解析代理字符串失败
			continue
		}

		ip := strings.TrimSpace(htmlquery.InnerText(ipTmp))
		port := strings.TrimSpace(htmlquery.InnerText(portTmp))

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

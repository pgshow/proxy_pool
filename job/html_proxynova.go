package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type proxyNova struct {
	Spider
}

func (s *proxyNova) StartUrl() []string {
	return []string{
		"https://www.proxynova.com/proxy-server-list/",
	}
}

func (s *proxyNova) Protocol() string {
	return "Fetch"
}

func (s *proxyNova) Cron() string {
	return "@every 2m"
}

func (s *proxyNova) GetReferer() string {
	return "https://www.proxynova.com/proxy-server-list/"
}

func (s *proxyNova) Run() {
	getProxy(s)
}

func (s *proxyNova) Name() string {
	return "proxyNova"
}

func (s *proxyNova) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@id='tbl_proxy_list']/tbody/tr")
	for _, n := range list {
		ipTmp := htmlquery.FindOne(n, "//td[1]/abbr/@title")
		portTmp := htmlquery.FindOne(n, "//td[2]")

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

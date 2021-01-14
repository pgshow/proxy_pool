package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type freeProxyListNet struct {
	Spider
}

func (s *freeProxyListNet) StartUrl() []string {
	return []string{
		"https://free-proxy-list.net/",
	}
}

func (s *freeProxyListNet) Protocol() string {
	return "Fetch"
}

func (s *freeProxyListNet) Cron() string {
	return "@every 10m"
}

func (s *freeProxyListNet) GetReferer() string {
	return "https://free-proxy-list.net/"
}

func (s *freeProxyListNet) Run() {
	getProxy(s)
}

func (s *freeProxyListNet) Name() string {
	return "freeProxyListNet"
}

func (s *freeProxyListNet) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='table table-striped table-bordered']/tbody/tr")
	for _, n := range list {
		ipTmp := htmlquery.FindOne(n, "//td[1]")
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

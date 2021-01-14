package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type freeProxyListsNet struct {
	Spider
}

func (s *freeProxyListsNet) StartUrl() []string {
	return []string{
		"http://www.freeproxylist.net/",
	}
}

func (s *freeProxyListsNet) Cron() string {
	return "@every 30m"
}

func (s *freeProxyListsNet) GetReferer() string {
	return "http://www.freeproxylist.net/"
}

func (s *freeProxyListsNet) Run() {
	getProxy(s)
}

func (s *freeProxyListsNet) Name() string {
	return "freeProxyListNet"
}

func (s *freeProxyListsNet) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='DataGrid']/tbody/tr[position()>1]")
	for _, n := range list {
		ipTmp := htmlquery.FindOne(n, "//a")
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

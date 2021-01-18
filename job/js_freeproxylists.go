package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type freeProxyLists struct {
	Spider
}

func (s *freeProxyLists) StartUrl() []string {
	return []string{
		"http://www.freeproxylist.net/",
	}
}

func (s *freeProxyLists) Profile() *Setting {
	return &Setting{
		Protocol:    "RenderFetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *freeProxyLists) Cron() string {
	return "@every 30m"
}

func (s *freeProxyLists) GetReferer() string {
	return "http://www.freeproxylist.net/"
}

func (s *freeProxyLists) Run() {
	getProxy(s)
}

func (s *freeProxyLists) Name() string {
	return "freeProxyList"
}

func (s *freeProxyLists) Parse(body string) (proxies []*model.HttpProxy, err error) {
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

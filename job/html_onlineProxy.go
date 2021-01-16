package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type onlineProxy struct {
	Spider
}

func (s *onlineProxy) StartUrl() []string {
	return []string{
		"http://online-proxy.ru/",
	}
}

func (s *onlineProxy) Protocol() string {
	return "Fetch"
}

func (s *onlineProxy) Cron() string {
	return "@every 24h"
}

func (s *onlineProxy) GetReferer() string {
	return "http://online-proxy.ru"
}

func (s *onlineProxy) Run() {
	getProxy(s)
}

func (s *onlineProxy) Name() string {
	return "onlineProxy"
}

func (s *onlineProxy) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='main']/tbody/tr[2]/td/table[3]/tbody/tr/td[4]/table[1]/tbody/tr[position()>1]")
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

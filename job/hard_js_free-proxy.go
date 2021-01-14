package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"strings"
)

type freeProxyCz struct {
	Spider
}

func (s *freeProxyCz) StartUrl() []string {
	return []string{
		"http://free-proxy.cz/en/",
	}
}

func (s *freeProxyCz) Protocol() string {
	return "Fetch"
}

func (s *freeProxyCz) Cron() string {
	return "@every 30m"
}

func (s *freeProxyCz) GetReferer() string {
	return "http://free-proxy.cz"
}

func (s *freeProxyCz) Run() {
	getProxy(s)
}

func (s *freeProxyCz) Name() string {
	return "freeProxy.Cz"
}

func (s *freeProxyCz) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@id='proxy_list']/tbody/tr")
	for _, n := range list {

		// 丢弃 Socks 协议
		if strings.Contains(htmlquery.OutputHTML(n, false), "SOCKS") {
			continue
		}

		ip := util.FindIp(htmlquery.OutputHTML(n, false))
		port := htmlquery.InnerText(htmlquery.FindOne(n, "//span[@class='fport']"))

		ip = strings.TrimSpace(ip)
		port = strings.TrimSpace(port)

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

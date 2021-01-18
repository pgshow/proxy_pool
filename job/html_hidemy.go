package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type hideMy struct {
	Spider
}

func (s *hideMy) StartUrl() []string {
	return []string{
		"https://hidemy.name/en/proxy-list/?type=hs#list",
	}
}

func (s *hideMy) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *hideMy) Cron() string {
	return "@every 5m"
}

func (s *hideMy) GetReferer() string {
	return "https://hidemy.name/en/proxy-list/?type=hs"
}

func (s *hideMy) Run() {
	getProxy(s)
}

func (s *hideMy) Name() string {
	return "hideMy"
}

func (s *hideMy) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//div[@class='table_block']/table/tbody/tr")
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

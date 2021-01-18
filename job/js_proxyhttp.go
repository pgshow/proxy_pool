package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"strings"
)

type proxyHttp struct {
	Spider
}

func (s *proxyHttp) StartUrl() []string {
	return []string{
		"https://proxyhttp.net/",
	}
}

func (s *proxyHttp) Profile() *Setting {
	return &Setting{
		Protocol:    "RenderFetch",
		AlwaysProxy: false,
		CnWebsite:   false,
	}
}

func (s *proxyHttp) Cron() string {
	return "@every 3m"
}

func (s *proxyHttp) GetReferer() string {
	return "https://proxyhttp.net/"
}

func (s *proxyHttp) Run() {
	getProxy(s)
}

func (s *proxyHttp) Name() string {
	return "proxyHttp"
}

func (s *proxyHttp) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='proxytbl']/tbody/tr[position()>1]")
	for _, n := range list {
		tmp := htmlquery.FindOne(n, "//td[@class='t_check']/a/@href")

		if tmp == nil {
			// 解析代理字符串失败
			continue
		}

		ipPort := util.FindIpPort(htmlquery.InnerText(tmp))

		ip := ipPort[0][1]
		port := ipPort[0][2]

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

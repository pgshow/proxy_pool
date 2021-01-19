package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type yip7 struct {
	Spider
}

func (s *yip7) StartUrl() []string {
	return []string{
		"https://www.7yip.cn/free/?action=china&page=1",
		"https://www.7yip.cn/free/?action=china&page=2",
		"https://www.7yip.cn/free/?action=china&page=3",
	}
}

func (s *yip7) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   true,
	}
}

func (s *yip7) Cron() string {
	return "@every 30m"
}

func (s *yip7) GetReferer() string {
	return "https://www.7yip.cn"
}

func (s *yip7) Run() {
	getProxy(s)
}

func (s *yip7) Name() string {
	return "yip7"
}

func (s *yip7) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table/tbody/tr[position()>1]")
	for _, n := range list {
		ip := htmlquery.InnerText(htmlquery.FindOne(n, "//td[1]"))
		port := htmlquery.InnerText(htmlquery.FindOne(n, "//td[2]"))

		ip = strings.TrimSpace(ip)
		port = strings.TrimSpace(port)

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

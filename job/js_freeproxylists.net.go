package job

import (
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"strings"
)

type freeproxylists struct {
	Spider
}

func (s *freeproxylists) StartUrl() []string {
	return []string{
		"http://www.freeproxylists.net/",
	}
}

func (s *freeproxylists) Cron() string {
	return "@every 30m"
}

func (s *freeproxylists) GetReferer() string {
	return "http://www.freeproxylists.net/"
}

func (s *freeproxylists) Run() {
	getProxy(s)
}

func (s *freeproxylists) Name() string {
	return "freeproxylists"
}

func (s *freeproxylists) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table[@class='DataGrid']/tbody/tr[position()>1]")
	for _, n := range list {
		ip := htmlquery.InnerText(htmlquery.FindOne(n, "//a/text()"))
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

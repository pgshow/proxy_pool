package job

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/util"
	"strings"
)

type xiladaili struct {
	Spider
}

func (s *xiladaili) StartUrl() []string {
	var u []string
	for _, d := range []string{"gaoni", "http", "https", "putong"} {
		for i := 1; i < 5; i++ {
			u = append(u, fmt.Sprintf("http://www.xiladaili.com/%s/%d/", d, i))
		}
	}
	return u
}

func (s *xiladaili) Profile() *Setting {
	return &Setting{
		Protocol:    "Fetch",
		AlwaysProxy: false,
		CnWebsite:   true,
	}
}

func (s *xiladaili) Cron() string {
	return "@every 2m"
}

func (s *xiladaili) GetReferer() string {
	return "http://www.xiladaili.com/"
}

func (s *xiladaili) Run() {
	getProxy(s)
}

func (s *xiladaili) Name() string {
	return "xiladaili"
}

func (s *xiladaili) Parse(body string) (proxies []*model.HttpProxy, err error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	list := htmlquery.Find(doc, "//table/tbody/tr[position()>1]")
	for _, n := range list {
		tmp := htmlquery.FindOne(n, "//td[1]")

		if tmp == nil {
			// 解析代理字符串失败
			continue
		}

		ip, port := util.Parse(strings.TrimSpace(htmlquery.InnerText(tmp)))

		proxies = append(proxies, &model.HttpProxy{
			Ip:   ip,
			Port: port,
		})
	}
	return
}

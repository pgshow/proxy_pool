package job

import (
	"encoding/json"
	"github.com/apex/log"
	"github.com/phpgao/proxy_pool/model"
)

type proxy11 struct {
	Spider
}

func (s *proxy11) StartUrl() []string {
	return []string{
		"https://proxy11.com/api/demoweb/proxy.json",
		"https://proxy11.com/api/demoweb/proxy.json?type=1",
	}
}

func (s *proxy11) Protocol() string {
	return "GET"
}

func (s *proxy11) GetReferer() string {
	return "https://proxy11.com/free-proxy"
}

func (s *proxy11) Run() {
	getProxy(s)
}

func (s *proxy11) Cron() string {
	return "@every 5m"
}

func (s *proxy11) Name() string {
	return "proxy11"
}

func (s *proxy11) TimeOut() int {
	return 60
}

type Item struct {
	Country      string  `json:"country"`
	Country_code string  `json:"country_code"`
	CreatedAt    string  `json:"createdAt"`
	Google       int     `json:"google"`
	Ip           string  `json:"ip"`
	Port         string  `json:"port"`
	Time         float64 `json:"time"`
	Type         int     `json:"type"`
	UpdatedAt    string  `json:"updatedAt"`
}

type proxy11J struct {
	Data []Item `json:"data"`
}

func (s *proxy11) Parse(body string) (proxies []*model.HttpProxy, err error) {
	var dataJson proxy11J
	err = json.Unmarshal([]byte(body), &dataJson)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"body":    body,
			"timeout": s.TimeOut(),
		}).Debug("error parse json")
		return
	}

	if len(dataJson.Data) == 0 {
		return
	}

	list := dataJson.Data

	for _, proxy := range list {
		proxies = append(proxies, &model.HttpProxy{
			Ip:   proxy.Ip,
			Port: proxy.Port,
		})
	}
	return
}

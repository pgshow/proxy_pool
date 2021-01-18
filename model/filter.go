package model

import (
	"github.com/phpgao/proxy_pool/ipdb"
	"github.com/phpgao/proxy_pool/util"
	"net"
	"strconv"
	"strings"
)

func filterOfSchema(v string) func(HttpProxy) bool {
	return func(proxy HttpProxy) bool {
		return proxy.Schema == strings.ToLower(v)
	}
}

func filterOfCn(v string) func(HttpProxy) bool {
	return func(proxy HttpProxy) bool {

		if strings.HasPrefix(v, "-") {
			// -号国家名（例如-cn）,代表剔除模式
			if "-"+proxy.Country != strings.ToLower(v) {
				return true
			}
		} else {
			// 非-号则是添加模式
			if proxy.Country == strings.ToLower(v) {
				return true
			}
		}

		return false
	}
}
func filterOfScore(v int) func(HttpProxy) bool {
	return func(proxy HttpProxy) bool {
		return proxy.Score >= v
	}
}
func filterOfSource(v string) func(HttpProxy) bool {
	return func(proxy HttpProxy) bool {
		return strings.EqualFold(proxy.From, v)
	}
}

func GetNewFilter(options map[string]string) (f []func(HttpProxy) bool, err error) {
	for k, v := range options {
		if k == "schema" && v != "" {
			f = append(f, filterOfSchema(v))
		}
		if k == "source" && v != "" {
			f = append(f, filterOfSource(v))
		}
		if k == "score" && v != "" {
			i := 0
			i, numError := strconv.Atoi(v)
			if numError != nil {
				if _, ok := numError.(*strconv.NumError); ok {
					i = 0
				} else {
					err = numError
					return
				}
			}
			f = append(f, filterOfScore(i))
		}

		if k == "country" && v != "" {
			f = append(f, filterOfCn(v))
		}
	}

	return
}

func FilterProxy(proxy *HttpProxy) bool {
	if tmp := net.ParseIP(proxy.Ip); tmp.To4() == nil {
		return false
	}

	port, err := strconv.Atoi(proxy.Port)
	if err != nil {
		return false
	}

	if port < 1 || port > 65535 {
		return false
	}

	ipInfo, err := ipdb.Db.FindInfo(proxy.Ip, "CN")
	if err != nil {
		logger.WithField("ip", proxy.Ip).WithError(err).Warn("can not find ip info")
		return false
	}

	if ipInfo.CountryName == "中国" {
		proxy.Country = "cn"
	} else {
		if util.ServerConf.OnlyChina {
			return false
		}
		proxy.Country = ipInfo.CountryName
	}

	return true
}

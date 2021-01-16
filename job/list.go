package job

import (
	"github.com/phpgao/proxy_pool/model"
)

var ListOfSpider = []Crawler{
	&proxyServers{},
	&proxyNova{},
	&onlineProxy{},
	&proxyHttp{},
	&proxyScan{},
	&proxyServers{},
	&proxyScrape{},
	&hideMy{},
	&rudnkh{},
	&proxy11{},
	&proxySale{},
	//&xici{},
	//&spys{},
	&pubProxy{},
	&kuaiProxy{},
	&cn66{},
	&hideMy{},
	//--&feiyi{},
	&ip89{},
	&goubanjia{},
	&ab57{},
	&clarketm{},
	&httptunnel{},
	&proxylist{},
	&proxylistplus{},
	&freeProxyListNet{},
	//--&aliveProxy{},
	&proxyDb{},
	&usProxy{},
	//---&siteDigger{},
	&dogdev{},
	&newProxy{},
	&xseo{},
	&ultraProxies{},
	&premProxy{},
	&nntime{},
	&proxyListsLine{},
	&myProxy{},
	&proxyIpList{},
	&blackHat{},
	&proxyLists{},
	&ip3366{},
	&xiladaili{},
	&nimadaili{},
	&openProxy{},
	//---&zdy{},
}

func GetSpiders(ch chan<- *model.HttpProxy) []Crawler {
	for _, v := range ListOfSpider {
		v.SetProxyChan(ch)
	}
	return ListOfSpider
}

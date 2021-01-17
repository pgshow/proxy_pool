package validator

import (
	"github.com/phpgao/proxy_pool/db"
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/queue"
	"github.com/phpgao/proxy_pool/util"
	"sync"
)

var (
	config      = util.ServerConf
	storeEngine = db.GetDb()
	lockMap     = sync.Map{}
)

func NewValidator() {
	q := queue.GetNewChan()
	var wg sync.WaitGroup
	for i := 0; i < config.OldQueue; i++ {
		wg.Add(1)
		go func() {
			for {
				proxy := <-q

				func(p *model.HttpProxy) {
					var err error
					logger := util.GetLogger("validator_new").WithField("from", p.From)
					key := p.GetKey()
					if _, ok := lockMap.Load(key); ok {
						return
					}

					lockMap.Store(key, 1)
					defer func() {
						lockMap.Delete(key)
					}()

					if storeEngine.Exists(*proxy) {
						return
					}

					err = p.SynTcpTest()
					if err != nil {
						logger.WithError(err).WithField("proxy", p.GetProxyUrl()).Debug("error test syn tcp")
						return
					}
					// http test
					err = p.TestProxy(false)
					if err != nil {
						logger.WithError(err).WithField(
							"proxy", p.GetProxyUrl()).Debug("error test http proxy")
						return
					} else {
						// https test
						err := p.TestProxy(true)
						if err != nil {
							logger.WithError(err).WithField(
								"proxy", p.GetProxyUrl()).Debug("error test https proxy")
						}
					}
					logger.WithField("proxy", p.GetProxyUrl()).Debug("add new proxy")
					storeEngine.Add(*p)
				}(proxy)
			}
		}()
	}
	wg.Wait()
}

package validator

import (
	"github.com/phpgao/proxy_pool/model"
	"github.com/phpgao/proxy_pool/queue"
	"github.com/phpgao/proxy_pool/util"
	"sync"
)

func OldValidator() {
	q := queue.GetOldChan()
	var wg sync.WaitGroup
	logger := util.GetLogger("validator_old")

	for i := 0; i < config.OldQueue; i++ {
		wg.Add(1)
		go func() {
			for {
				proxy := <-q
				func(p model.HttpProxy) {
					key := p.GetKey()
					if _, ok := lockMap.Load(key); ok {
						return
					}

					lockMap.Store(key, 1)
					defer func() {
						lockMap.Delete(key)
					}()
					if storeEngine.Exists(p) {
						var score int
						conn, err := p.TestTcp()
						if err != nil {
							score = -30
						}
						// if tcp test success
						if conn != nil {
							score = 10
							// test https

							err := p.TestConnectMethod(conn)
							if err != nil {
								logger.WithError(err).WithField("proxy", p.GetProxyWithSchema()).Debug("error https retest")
								if p.IsHttps() {
									score = -20
								}
							} else {
								p.Schema = "https"
								// save proxy to db
								err = storeEngine.UpdateSchema(p)
								if err != nil {
									logger.WithError(err).WithField("proxy", p.GetProxyWithSchema()).Info("error update schema")
								}
							}
						}
						//logger.WithFields(log.Fields{
						//	"score": score,
						//	"proxy": p.GetProxyWithSchema(),
						//}).Debug("set score")

						err = storeEngine.AddScore(p, score)
						if err != nil {
							logger.WithError(err).WithField(
								"proxy", p.GetProxyWithSchema()).Error("error setting score")
						}
					}
				}(*proxy)
			}

		}()

	}
	wg.Wait()
}

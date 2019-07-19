package binance

import (
	log "github.com/sirupsen/logrus"
	dic "msq.ai/db/postgres/dictionaries"
)

func RunBinanceConnector(dictionaries *dic.Dictionaries, apiKey string, secretKey string) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

	ctxLog.Info("BinanceConnector is going to start")

}

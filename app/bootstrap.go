package app

import (
	"megin/app/consumer"
	"megin/library/logger"
)

// 此方法在mysql,redis等连接完成后,gin运行前调用
// 你可以在此处初始化业务相关内容，比如初始化全局变量等
func OnAppInitialize() error {
	logger.Info("OnAppInitialize Run....")
	initListeningPort()
	return nil
}

func initMqttConsumer() {

}

//func initKafkaConsumer() {
//	//go consumer.Consumer()
//	go consumer.BatchConsumer()
//}

func initListeningPort() {

	go consumer.BatchHost()
}

package consumer

import (
	"fmt"
	"megin/system/config"
	"strings"
)

const (
	TopicGameOrder    = "topic_yuanbao_${env}_game_orders"
	TopicWalletChange = "topic_yuanbao_${env}_wallet_change"

	TopicWalletChangeBinlogPH = "mysqlbinlog_yuanbao_ph_wallet"
	TopicGameTicketBinlogPH   = "mysqlbinlog_yuanbao_ph_game_ticket"

	TopicWalletChangeBinlogVN = "mysqlbinlog_yuanbao_wallet"
)

const (
	GroupDefault          = "group_elastic_default" //默认分组
	GroupElasticGameOrder = "group_elastic_game_order"
)

type GroupTopicConfig struct {
}

// 不同环境topic名称不一样
func getEnvTopics(conf *config.ServiceConfig, topics []string) []string {
	for i, topic := range topics {
		if conf.Env == config.EnvDev || conf.Env == config.EnvTest {
			topics[i] = strings.ReplaceAll(topic, "${env}", "test")
		} else {
			topics[i] = strings.ReplaceAll(topic, "${env}_", "")
		}
	}
	return topics
}

func getEnvTopic(conf *config.ServiceConfig, topic string) string {
	if conf.Env == config.EnvDev || conf.Env == config.EnvTest {
		topic = strings.ReplaceAll(topic, "${env}", "test")
	} else {
		topic = strings.ReplaceAll(topic, "${env}_", "")
	}
	return topic
}

func BatchHost() {
	conf := config.GetConfig()
	if len(conf.Listening.Host+":"+conf.Listening.Port) == 0 {
		return
	}

	fmt.Println("监听接收机、发射机端口：" + conf.Listening.Host + ":" + conf.Listening.Port + "获取成功")

	go startListeningPortReceiver(conf.Listening.Host, conf.Listening.Port)

	go StartWsServer("8900")
}

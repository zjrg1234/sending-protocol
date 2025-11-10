package kafka_test

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang-module/carbon/v2"
	consumer2 "megin/app/consumer"
	"megin/library/kafka/consumer"
	"megin/library/kafka/producer"
	"megin/library/logger"
	"megin/system/config"
	"testing"
	"time"
)

func TestSendMessage(t *testing.T) {
	go ConsumerMessage()
	time.Sleep(time.Second * 2)
	for i := 0; i < 10000; i++ {
		now := carbon.Now().ToDateTimeString()
		msg := fmt.Sprintf("我是消息_%s_%d", now, i)
		producer.PushMessage("test", msg)
		//	fmt.Println("发送消息:" + msg)
		time.Sleep(time.Millisecond * 100)
	}
}

func TestConsumerMessage(t *testing.T) {
	go ConsumerMessage()
	time.Sleep(time.Second * 200)
}

func TestSendBatchMessage(t *testing.T) {
	go ConsumerBatchMessage()

	batchId := carbon.Now().Timestamp()
	time.Sleep(time.Second * 2)
	for i := 0; i < 10000; i++ {
		msg := fmt.Sprintf("我是消息_%d_%d", batchId, i)
		producer.PushMessage("test_batch_topic", msg)
		//	fmt.Println("发送消息:" + msg)
		time.Sleep(time.Millisecond * 100)
	}
}

//func(message []*ConsumerSessionMessage) error
func onBatchMessage(message []*consumer2.ConsumerSessionMessage) error {
	fmt.Println(message)
	return nil
}

func ConsumerBatchMessage() {
	conf := config.GetConfig()
	consumer2.StartBatchConsumer(conf.KafkaBroker, "test_batch_topic", "group_2", onBatchMessage)
}

func MessageHandle(message *sarama.ConsumerMessage) error {
	msg := fmt.Sprintf("接收消息: value = %s, timestamp = %v, topic = %s, partions = %d, offset = %d", string(message.Value), message.Timestamp, message.Topic, message.Partition, message.Offset)
	fmt.Println(msg)
	return nil
}

func ConsumerMessage() {
	conf := config.GetConfig()
	group := consumer.NewKafkaConsumerGroupAction([]string{conf.KafkaBroker}, "group_2")
	group.Consume([]string{"test"}, MessageHandle)
}

func TestMain(m *testing.M) {
	fmt.Println("run main")
	//解析配置文件
	config.InitConfig("../../resources/config-dev.yaml")
	//Log初始化
	logger.InitLog(logger.LogConfig{LogInConsole: true})
	m.Run()
}

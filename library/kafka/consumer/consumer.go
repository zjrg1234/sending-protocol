package consumer

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"megin/library/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func ShowMetadata(addrs []string) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return
	}
	defer client.Close()
	// get topic set
	topics, err := client.Topics()
	if err != nil {
		fmt.Printf("try get topics err %s\n", err.Error())
		return
	}

	fmt.Printf("topics(%d):\n", len(topics))
	for _, topic := range topics {
		logger.Info("ShowMetadata", zap.String("topic", topic))
	}

	// get broker set
	brokers := client.Brokers()
	fmt.Printf("broker set(%d):\n", len(brokers))
	for _, broker := range brokers {
		fmt.Printf("%s\n", broker.Addr())
	}
}

func MessageHandle(message *sarama.ConsumerMessage) {
	msg := fmt.Sprintf("Message claimed: value = %s, timestamp = %v, topic = %s, partions = %d, offset = %d",
		string(message.Value), message.Timestamp, message.Topic, message.Partition, message.Offset)
	logger.Info(msg)
}

type KafkaConsumerGroupAction struct {
	group  sarama.ConsumerGroup
	handle func()
}

func NewKafkaConsumerGroupAction(brokers []string, groupId string) *KafkaConsumerGroupAction {
	logger.Info("NewKafkaConsumerGroupAction", zap.String("groupId", groupId), zap.Any("brokers", brokers))
	ShowMetadata(brokers)
	config := sarama.NewConfig()
	sarama.Logger = log.New(os.Stdout, "[consumer_group]", log.Lshortfile)
	// 重平衡策略
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 6 * time.Second
	config.Consumer.IsolationLevel = sarama.ReadCommitted
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	//fetch.message.max.bytes
	config.Consumer.Fetch.Default = 52428800
	//fetch.min.bytes
	config.Consumer.Fetch.Min = 30000
	config.Consumer.MaxWaitTime = time.Second * 10

	//手动提交,判断是否异常
	//config.Consumer.Offsets.AutoCommit.Enable = false
	config.Version = sarama.V2_7_0_0
	consumerGroup, e := sarama.NewConsumerGroup(brokers, groupId, config)

	if e != nil {
		logger.Info("NewKafkaConsumerGroupAction", zap.Error(e))
		return nil
	}
	return &KafkaConsumerGroupAction{group: consumerGroup}
}

func (K *KafkaConsumerGroupAction) Consume(topics []string, handleFunc func(message *sarama.ConsumerMessage) error) {

	if K == nil {
		return
	}
	var wg sync.WaitGroup
	ctx := context.Background()
	var consumer = KafkaConsumerGroupHandler{ready: make(chan bool), handle: handleFunc}
	go func() {
		defer wg.Done()
		for {
			if err := K.group.Consume(ctx, topics, &consumer); err != nil {
				logger.Info("consumer|Error from consumer", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	logger.Info("Sarama consumer up and running!...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		logger.Info("consumer|terminating: context cancelled")
	case <-sigterm:
		logger.Info("consumer|terminating: via signal")
	}
	wg.Wait()
	if err := K.group.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

type KafkaConsumerGroupHandler struct {
	ready  chan bool
	handle func(message *sarama.ConsumerMessage) error
}

func (consumer *KafkaConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *KafkaConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *KafkaConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		//处理失败,判断是否需要提交
		err := consumer.handle(message)

		if rand.Int31n(50) == 0 {
			count := claim.HighWaterMarkOffset() - message.Offset
			logger.Info("有N条消息待消费", zap.String("message.Topic", message.Topic), zap.Int64("count", count))
		}
		session.MarkMessage(message, "")

		if err != nil {
			//TODO::判断err
		}

		//手动提交时,提交
		//session.Commit()

	}
	return nil
}

func (consumer *KafkaConsumerGroupHandler) BatchConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//TODO:: 待实现在批量消费, 批量+ 超时控制
	//https://blog.csdn.net/kenkao/article/details/121033727

	for message := range claim.Messages() {
		//处理失败,判断是否需要提交
		err := consumer.handle(message)

		if rand.Int31n(500) == 0 {
			lag := claim.HighWaterMarkOffset() - message.Offset
			logger.Info("有N条消息待消费", zap.Int64("lag", lag))
		}
		session.MarkMessage(message, "")

		if err != nil {
			//TODO::判断err
		}

		//手动提交时,提交
		session.Commit()

	}
	return nil
}

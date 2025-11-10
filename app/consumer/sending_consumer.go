package consumer

import (
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"log"
	"megin/app/repo"
	"megin/library/logger"
	"megin/library/redis"
	"net"
	"strconv"
)

func startForwardingReceiver(transmitterId string, rawData []byte, clientAddrStr string) {
	receiverIdRedisKey := transmitterId //取到绑定的receiver
	receiverId := redis.Get(receiverIdRedisKey).Val()

	receiverRedisKey := string(receiverId) + "_receiver_host_port" //端口
	receiverHostPort := redis.Get(receiverRedisKey).Val()

	transmitterRedisKey := string(transmitterId) + "_transmitter_host_port" //端口
	transmitterHostPort := redis.Get(transmitterRedisKey)
	fmt.Println("取出redis数据:", transmitterHostPort.Val())
	fmt.Println("key:", transmitterRedisKey)
	if transmitterHostPort.Val() == "" {
		err := redis.Set(transmitterRedisKey, clientAddrStr, 5)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}
	log.Println("解析地址receiverHostPort：", receiverHostPort)

	serverAddr, err := net.ResolveUDPAddr("udp", receiverHostPort)
	log.Println("解析地址：", serverAddr)

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Printf("连接 %s 失败：%v", serverAddr.String(), err)
		return
	}
	_, err = conn.Write(rawData)
	if err != nil {
		log.Printf("回复客户端 %s 失败：%v", serverAddr.String(), err)
		return
	}

}

func getReceiverMessage(receiverId string, rawData []byte, clientAddrStr string) {
	supplyVoltageByte := rawData[30:32]
	//处理配置表
	vehicleConfig, err := repo.GetVehicleConfig(receiverId) //车辆
	if err != nil {
		logger.Error("查询车辆配置失败:", zap.Error(err))
		return
	}
	vehicleConfig.ReceiverHostPort = clientAddrStr

	err = repo.UpdateVehicleConfig(vehicleConfig)
	if err != nil {
		logger.Error("更新车辆配置错误 :", zap.Error(err))
		return
	}
	vehicle, err := repo.GetVehicle(receiverId) //车辆
	if err != nil {
		logger.Error("查询车辆失败:", zap.Error(err))
		return
	}
	//处理车辆
	supplyVoltageStr := string(supplyVoltageByte)
	supplyVoltageTen, err := hex.DecodeString(supplyVoltageStr) //车辆id或发射机id
	batter := float64(supplyVoltageTen[0]) / 10.0
	vehicle.VehicleBattery = strconv.FormatFloat(batter, 'f', 1, 64)
	err = repo.UpdateVehicle(vehicle) //车辆
	if err != nil {
		logger.Error("更新车辆失败:", zap.Error(err))
		return
	}
	redisKey := string(receiverId) + "_receiver_host_port" //端口
	test := redis.Get(redisKey)
	fmt.Println("取出redis数据:", test.Val())
	fmt.Println("key:", redisKey)
	if test.Val() == "" {
		err = redis.Set(redisKey, clientAddrStr, 0)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}
}

func getReceiverHeartBeat(receiverId string, rawData []byte, clientAddrStr string) {
	transmitterIdRedisKey := string(receiverId) //取到绑定的transmitter
	transmitterId := redis.Get(transmitterIdRedisKey).Val()

	transmitterRedisKey := string(transmitterId) + "_transmitter_host_port" //端口
	transmitterHostPort := redis.Get(transmitterRedisKey).Val()

	receiverRedisKey := string(receiverId) + "_receiver_host_port" //端口
	receiverHostPort := redis.Get(receiverRedisKey)

	fmt.Println("取出redis数据:", receiverHostPort.Val())
	fmt.Println("key:", receiverRedisKey)
	if receiverHostPort.Val() == "" {
		err := redis.Set(receiverRedisKey, clientAddrStr, 0)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}

	serverAddr, err := net.ResolveUDPAddr("udp", transmitterHostPort) //发送
	conn, err := net.DialUDP("udp", nil, serverAddr)

	_, err = conn.Write(rawData)
	if err != nil {
		log.Printf("回复发送端 %s 失败：%v", serverAddr.String(), err)
		return
	}
}

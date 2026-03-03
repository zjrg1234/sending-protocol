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
	"strings"
)

type ForwardServer struct {
	conn     *net.UDPConn
	port     string
	msgCount int64 // 消息计数器，用于监控
}

// 监听
func startListeningPortReceiver(host string, port string) {

	heartBeatPort := ":8898" //接收机接收心跳ip
	log.Printf("多对多转发服务器启动在端口: %d (使用Redis存储)", port)
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	// 4. 创建 UDP 监听器
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Error("UDP 监听失败：", zap.Error(err))
		return
	}
	//方便后续调用
	server := &ForwardServer{
		conn: conn,
		port: port,
	}
	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("读取 UDP 数据失败: %v\n", err)
			continue
		}

		fmt.Printf("发送方ip加端口: %q\n", clientAddr.String()) //测试使用后期注释
		//if clientAddr.IP.String() == "127.0.0.1" { //本地测试
		//	rawData := buffer[:n]
		//	//1. 字节数组 → 十六进制字符串（默认小写）
		//	//hexLower := hex.EncodeToString(hexRawData)
		//	//转大写（更易读，协议常用）
		//	//rawData := strings.ToUpper(hexLower)
		//	fmt.Printf("\n收到来自 %s 的 UDP 数据（原始）: %s\n", clientAddr, rawData)
		//	fmt.Println("原始十进制数据：", rawData, "结束\n")
		//	header := rawData[0:4]
		//	if string(header) != "5A43" {
		//		continue
		//	}
		//	hexStr := string(rawData[10:26])
		//
		//	id, err := hex.DecodeString(hexStr) //车辆或发射机id
		//	if err != nil {
		//		continue
		//	}
		//	commandCode := string(rawData[4:6]) //命令码
		//
		//	if commandCode == "10" {
		//		go startForwardingReceiver(server, string(id), rawData, clientAddr)
		//	}
		//	if commandCode == "15" {
		//		go getReceiverMessage(server, string(id), rawData, clientAddr)
		//	}
		//	if commandCode == "16" {
		//		go getReceiverHeartBeat(server, string(id), rawData, clientAddr, heartBeatPort)
		//	}
		//	fmt.Printf("获取到车辆id或发射机id: %q\n", id)
		//	fmt.Printf("命令码: %q\n", commandCode)
		//} else {
		// 原始数据
		hexRawData := buffer[:n]
		// 1. 字节数组 → 十六进制字符串（默认小写）
		hexLower := hex.EncodeToString(hexRawData)
		// 转大写（更易读，协议常用）
		rawData := strings.ToUpper(hexLower)
		fmt.Printf("\n收到来自 %s 的 UDP 数据（原始）: %s\n", clientAddr, hexRawData)
		fmt.Println("原始十进制数据：", hexRawData, "结束\n")
		header := rawData[0:4]
		if string(header) != "5A43" {
			continue
		}
		hexStr := rawData[10:26]
		id, err := hex.DecodeString(hexStr) //车辆或发射机id
		if err != nil {
			continue
		}
		commandCode := rawData[4:6] //命令码

		if commandCode == "10" {
			go startForwardingReceiver(server, string(id), hexRawData, clientAddr)
		}
		if commandCode == "15" {
			go getReceiverMessage(server, string(id), hexRawData, clientAddr)
		}
		if commandCode == "16" {
			go getReceiverHeartBeat(server, string(id), hexRawData, clientAddr, heartBeatPort)
		}
		fmt.Printf("获取到车辆id或发射机id: %q\n", id)
		fmt.Printf("命令码: %q\n", commandCode)
		//}

	}

}

// 发射机转发
func startForwardingReceiver(server *ForwardServer, transmitterId string, rawData []byte, clientAddrStr *net.UDPAddr) {
	receiverIdRedisKey := transmitterId //取到绑定的receiver
	receiverId := redis.Get(receiverIdRedisKey).Val()

	//transmitterRedisKey := string(transmitterId) + "_transmitter_host_port" //端口
	//transmitterHostPort := redis.Get(transmitterRedisKey)

	receiverRedisKey := string(receiverId) + "_receiver" //对应车辆配置信息 包含端口
	ClientInfo, err := redis.GetClientInfo(receiverRedisKey)
	if err != nil {
		logger.Info("未获取到redis缓存:" + receiverRedisKey)
		logger.Error("未获取到redis缓存:", zap.Error(err))
		return
	}
	ClientInfo.TransmitterId = transmitterId

	fmt.Println(transmitterId, "取出ReceiverHost数据:", ClientInfo.ReceiverHostPort)
	//clientAddrStrSet := clientAddrStr.IP.String() + ":" + "8898"
	clientAddrStrSet := clientAddrStr.String()

	if ClientInfo.TransmitterHostPort != clientAddrStrSet {
		ClientInfo.TransmitterHostPort = clientAddrStrSet
		err := redis.SaveClientInfo(receiverRedisKey, ClientInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}

	serverAddr, err := net.ResolveUDPAddr("udp", ClientInfo.ReceiverHostPort)
	if err != nil {
		log.Printf("连接 %s 失败：%v", err)
		return
	}
	_, err = server.conn.WriteToUDP(rawData, serverAddr)
	if err != nil {
		log.Printf("发送消息到 %s 失败: %v", serverAddr.String(), err)
		return
	} else {
		log.Printf("回复发送端 %s 成功", serverAddr.String())
		return
	}

}

// 开机
func getReceiverMessage(server *ForwardServer, receiverId string, rawData []byte, clientAddrStr *net.UDPAddr) {
	//supplyVoltageByte := rawData[30:32]
	//处理配置表
	vehicleConfig, err := repo.GetVehicleConfig(receiverId) //车辆
	if err != nil {
		logger.Error("查询车辆配置失败:", zap.Error(err))
		return
	}
	vehicleConfig.ReceiverHostPort = clientAddrStr.String()

	err = repo.UpdateVehicleConfig(vehicleConfig)
	if err != nil {
		logger.Error("更新车辆配置错误 :", zap.Error(err))
		return
	}
	//vehicle, err := repo.GetVehicle(receiverId) //车辆
	//if err != nil {
	//	logger.Error("查询车辆失败:", zap.Error(err))
	//	return
	//}
	//处理车辆
	//supplyVoltageStr := string(supplyVoltageByte)
	//supplyVoltageTen, err := hex.DecodeString(supplyVoltageStr) //车辆id或发射机id
	//batter := float64(supplyVoltageTen[0]) / 10.0
	//vehicle.VehicleBattery = strconv.FormatFloat(batter, 'f', 1, 64)
	//err = repo.UpdateVehicle(vehicle) //车辆
	//if err != nil {
	//	logger.Error("更新车辆失败:", zap.Error(err))
	//	return
	//}
	redisKey := string(receiverId) + "_receiver" //端口
	clientInfo, err := redis.GetClientInfo(redisKey)

	fmt.Println("取出redis数据:", clientInfo)
	fmt.Println("key:", redisKey)
	if clientInfo == nil {
		clientSetInfo := &redis.ClientInfo{
			ReceiverId:          receiverId,
			ReceiverHostPort:    clientAddrStr.String(),
			TransmitterId:       "0",
			TransmitterHostPort: "",
		}
		err = redis.SaveClientInfo(redisKey, clientSetInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	} else {
		clientInfo.ReceiverId = receiverId
		clientInfo.ReceiverHostPort = clientAddrStr.String()
		err = redis.SaveClientInfo(redisKey, clientInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}
}

// 发射机心跳
func getReceiverHeartBeat(server *ForwardServer, receiverId string, rawData []byte, clientAddrStr *net.UDPAddr, heartBeatPort string) {

	receiverRedisKey := string(receiverId) + "_receiver" //取到绑定信息
	ClientInfo, err := redis.GetClientInfo(receiverRedisKey)
	if ClientInfo == nil {
		logger.Info("未获取到redis缓存:" + receiverRedisKey)
		return
	}

	ClientInfo.ReceiverId = receiverId

	fmt.Println(receiverId, "取出TransmitterHostPort数据:", ClientInfo.TransmitterHostPort)
	if ClientInfo.ReceiverHostPort != clientAddrStr.String() {
		ClientInfo.ReceiverHostPort = clientAddrStr.String()
		err := redis.SaveClientInfo(receiverRedisKey, ClientInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}

	serverAddr, err := net.ResolveUDPAddr("udp", ClientInfo.TransmitterHostPort) //发送
	log.Println("解析地址transmitterHostPort：", serverAddr)
	if err != nil {
		log.Printf("连接 %s 失败：%v", err)
		return
	}
	_, err = server.conn.WriteToUDP(rawData, serverAddr)
	if err != nil {
		log.Printf("回复发送端 %s 失败：%v", serverAddr.String(), err)
		return
	} else {
		log.Printf("回复发送端 %s 成功", serverAddr.String())
		return
	}
}

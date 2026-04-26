package consumer

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"megin/app/repo"
	"megin/library/logger"
	"megin/library/redis"
	"net"
	"sync"
	"time"
)

// ================= 全局变量与核心结构 =================

var (
	// SessionMap 存储 3 秒有效期的缓存，阻挡高频 Redis 查询
	SessionMap sync.Map
)

// HotCache 本地热缓存，用于 0.04s 极速转发
type HotCache struct {
	mu            sync.RWMutex
	ReceiverAddr  *net.UDPAddr // 接收机(车辆)的目标地址
	LastRedisSync time.Time    // 上次去 Redis 查数据的时间
}

type ForwardServer struct {
	conn     *net.UDPConn
	port     string
	msgCount int64 // 消息计数器，用于监控
}

// 监听
func startListeningPortReceiver(host string, port string) {

	heartBeatPort := ":8898" //接收机接收心跳ip
	log.Printf("多对多转发服务器启动在端口: %s (3秒TTL缓存 + 完整业务融合版)", port)
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
		if buffer[0] != 0x5A || buffer[1] != 0x43 {
			continue
		}
		commandCode := buffer[2]      //命令码
		idStr := string(buffer[5:13]) //id

		if commandCode == 0x10 {
			startForwardingReceiver(server, idStr, buffer[:n], clientAddr)
		}
		if commandCode == 0x15 {
			dataCopy := make([]byte, n)
			copy(dataCopy, buffer[:n])
			go getReceiverMessage(server, idStr, dataCopy, clientAddr)
		}
		if commandCode == 0x16 {
			dataCopy := make([]byte, n)
			copy(dataCopy, buffer[:n])
			go getReceiverHeartBeat(server, idStr, dataCopy, clientAddr, heartBeatPort)
		}
		if commandCode == 0x14 {
			dataCopy := make([]byte, n)
			copy(dataCopy, buffer[:n])
			go getTransmitterHeartBeat(server, idStr, dataCopy, clientAddr, heartBeatPort)
		}
		fmt.Printf("获取到车辆id或发射机id: %q\n", idStr)
		fmt.Printf("命令码: %q\n", commandCode)
		//}

	}

}

// 发射机转发
func startForwardingReceiver(server *ForwardServer, transmitterId string, rawData []byte, clientAddrStr *net.UDPAddr) {
	cacheIface, _ := SessionMap.LoadOrStore(transmitterId, &HotCache{})
	cache := cacheIface.(*HotCache)

	cache.mu.RLock()
	timeSinceLastSync := time.Since(cache.LastRedisSync)
	targetAddr := cache.ReceiverAddr
	cache.mu.RUnlock()

	if timeSinceLastSync > 2*time.Second {
		go func(tId string, currentClientAddr string) {
			receiverIdRedisKey := transmitterId //取到绑定的receiver
			receiverId := redis.Get(receiverIdRedisKey).Val()
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
			// 解析出车辆 IP 端口，并更新本地 3 秒热缓存
			newAddr, err := net.ResolveUDPAddr("udp", ClientInfo.ReceiverHostPort)
			if err == nil {
				cache.mu.Lock()
				cache.ReceiverAddr = newAddr
				cache.LastRedisSync = time.Now() // 重置 3 秒 TTL
				cache.mu.Unlock()
			}
			if newAddr != nil {
				_, err := server.conn.WriteToUDP(rawData, newAddr)
				if err != nil {
					// log.Printf("发送消息到 %s 失败: %v", targetAddr.String(), err)
				}
			}
		}(transmitterId, clientAddrStr.String())
		// 极速转发：不管刚才的 go func 查没查完，先用当前手里的地址把指令发给车辆！
		// 这是保证 0.04s (25Hz) 丝滑驾驶的关键！
	}
	if targetAddr != nil {
		_, err := server.conn.WriteToUDP(rawData, targetAddr)
		if err != nil {
			// log.Printf("发送消息到 %s 失败: %v", targetAddr.String(), err)
		}
	}

	//transmitterRedisKey := string(transmitterId) + "_transmitter_host_port" //端口
	//transmitterHostPort := redis.Get(transmitterRedisKey)

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
	//
	//err = repo.UpdateVehicleConfig(vehicleConfig)
	//if err != nil {
	//	logger.Error("更新车辆配置错误 :", zap.Error(err))
	//	return
	//}
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

// 接收机心跳
func getReceiverHeartBeat(server *ForwardServer, receiverId string, rawData []byte, clientAddrStr *net.UDPAddr, heartBeatPort string) {

	receiverRedisKey := string(receiverId) + "_receiver" //取到绑定信息
	ClientInfo, err := redis.GetClientInfo(receiverRedisKey)

	if ClientInfo == nil {
		clientSetInfo := &redis.ClientInfo{
			ReceiverId:          receiverId,
			ReceiverHostPort:    clientAddrStr.String(),
			TransmitterId:       "0",
			TransmitterHostPort: "",
		}
		err := redis.SaveClientInfo(receiverRedisKey, clientSetInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
		logger.Info("未获取到开机redis缓存 重新塞入:" + receiverRedisKey)

		return
	}
	ClientInfo.ReceiverId = receiverId

	fmt.Println(receiverId, "取出TransmitterHostPort数据:", ClientInfo.TransmitterHostPort)
	ClientInfo.ReceiverHostPort = clientAddrStr.String()

	//if ClientInfo.ReceiverHostPort != clientAddrStr.String() {
	ClientInfo.ReceiverHostPort = clientAddrStr.String()
	err = redis.SaveClientInfo(receiverRedisKey, ClientInfo)
	if err != nil {
		logger.Error("redis塞入错误:", zap.Error(err))
		return
	}
	//}

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

// 发射机心跳
func getTransmitterHeartBeat(server *ForwardServer, transmitterId string, rawData []byte, clientAddrStr *net.UDPAddr, heartBeatPort string) {

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

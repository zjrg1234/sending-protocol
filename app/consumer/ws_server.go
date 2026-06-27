package consumer

import (
	"log"
	"net"
	"net/http"
	"time"

	"megin/app/repo"
	"megin/library/logger"
	"megin/library/redis"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// ================= 全局与升级器配置 =================

var (
	WsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	wsLocalUdpConn *net.UDPConn
)

// ================= 核心服务启动入口 =================

func StartWsServer(port string) {
	var err error
	wsLocalUdpConn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		logger.Error("UDP 套接字初始化失败:", zap.Error(err))
		return
	}
	// 注意不要在此 defer wsLocalUdpConn.Close()，让它常驻后台发包

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ws", handleWebSocketPortal)

	bindAddr := "0.0.0.0:" + port
	logger.Info("WebSocket 服务启动在端口: " + port)

	if err := r.Run(bindAddr); err != nil {
		logger.Error("WebSocket 启动失败:", zap.Error(err))
	}
}

func handleWebSocketPortal(c *gin.Context) {
	wsConn, err := WsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer wsConn.Close()

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			break
		}

		if len(message) < 13 {
			continue
		}
		if message[0] != 0x5A || message[1] != 0x43 {
			continue
		}

		commandCode := message[2]
		idStr := string(message[5:13])

		dataCopy := make([]byte, len(message))
		copy(dataCopy, message)

		heartBeatPort := ":8898" // 保持和你原代码一致

		if commandCode == 0x10 {
			go handleWsForwardingTask(wsConn, idStr, dataCopy)
		}
		if commandCode == 0x15 {
			go handleWsReceiverMessage(wsConn, idStr, dataCopy)
		}
		if commandCode == 0x16 {
			go handleWsReceiverHeartBeat(wsConn, idStr, dataCopy, heartBeatPort, message[5:13])
		}
		if commandCode == 0x14 {
			go handleWsTransmitterHeartBeat(wsConn, idStr, dataCopy, heartBeatPort)
		}
	}
}

// =====================================================================
// 严格参照原版代码重写的 WS 业务处理集
// =====================================================================

// 发射机转发
func handleWsForwardingTask(wsConn *websocket.Conn, transmitterId string, rawData []byte) {
	clientAddrStr := wsConn.RemoteAddr().String()

	// 🚨 架构师防线 3：在入口处就地处决 APP 误发的 0x16 心跳！绝不浪费 CPU 和 Redis 资源
	if rawData[2] == 0x16 {
		logger.Warn("🚨 抓到内鬼！APP 端企图发送 0x16 心跳指令，已强行拦截！", zap.String("app_addr", clientAddrStr))
		return
	}
	cacheIface, _ := SessionMap.LoadOrStore(transmitterId, &HotCache{})
	cache := cacheIface.(*HotCache)

	cache.mu.RLock()
	timeSinceLastSync := time.Since(cache.LastRedisSync)
	targetAddr := cache.ReceiverAddr
	cache.mu.RUnlock()

	if targetAddr == nil || timeSinceLastSync > 2*time.Second {
		go func(tId string, currentClientAddr string, payload []byte) {
			receiverIdRedisKey := transmitterId //取到绑定的receiver
			receiverId := redis.Get(receiverIdRedisKey).Val()
			if receiverId == "" {
				log.Printf("未获取到receiverId缓存:")
				return
			}

			receiverRedisKey := string(receiverId) + "_receiver" //对应车辆配置信息 包含端口
			ClientInfo, err := redis.GetClientInfo(receiverRedisKey)
			if err != nil {
				logger.Error("未获取到redis缓存:", zap.Error(err))
				return
			}

			ClientInfo.TransmitterId = transmitterId

			clientAddrStrSet := currentClientAddr

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

				// 🚨 架构师防线 5：修复导致车子发抖的“幽灵双重重传”！
				if targetAddr == nil {
					_, err = safeWriteUDP(wsLocalUdpConn, payload, newAddr)
					if err != nil {
						log.Printf("发送消息到 %s 失败: %v", newAddr.String(), err)
					}
				}
			}

			if newAddr != nil {
				_, err := wsLocalUdpConn.WriteToUDP(rawData, newAddr)
				if err != nil {
					log.Printf("发送消息到 %s 失败: %v", targetAddr.String(), err)
				} else {
					return
				}
			}
		}(transmitterId, clientAddrStr, rawData)
	}
	if targetAddr != nil {
		_, err := safeWriteUDP(wsLocalUdpConn, rawData, targetAddr)
		if err != nil {
			log.Printf("发送消息到 %s 失败: %v", targetAddr.String(), err)
		} else {
			return
		}
	}
}

// 开机
func handleWsReceiverMessage(wsConn *websocket.Conn, receiverId string, rawData []byte) {
	clientAddrStr := wsConn.RemoteAddr().String()

	//处理配置表
	vehicleConfig, err := repo.GetVehicleConfig(receiverId) //车辆
	if err != nil {
		logger.Error("查询车辆配置失败:", zap.Error(err))
		return
	}
	vehicleConfig.ReceiverHostPort = clientAddrStr

	redisKey := string(receiverId) + "_receiver" //端口
	clientInfo, err := redis.GetClientInfo(redisKey)

	if clientInfo == nil {
		clientSetInfo := &redis.ClientInfo{
			ReceiverId:          receiverId,
			ReceiverHostPort:    clientAddrStr,
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
		clientInfo.ReceiverHostPort = clientAddrStr
		err = redis.SaveClientInfo(redisKey, clientInfo)
		if err != nil {
			logger.Error("redis塞入错误:", zap.Error(err))
			return
		}
	}
}

// 接收机心跳
func handleWsReceiverHeartBeat(wsConn *websocket.Conn, receiverId string, rawData []byte, heartBeatPort string, receiverBuff []byte) {
	clientAddrStr := wsConn.RemoteAddr().String()

	receiverRedisKey := string(receiverId) + "_receiver" //取到绑定信息
	ClientInfo, err := redis.GetClientInfo(receiverRedisKey)

	if ClientInfo == nil {
		clientSetInfo := &redis.ClientInfo{
			ReceiverId:          receiverId,
			ReceiverHostPort:    clientAddrStr,
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

	ClientInfo.ReceiverHostPort = clientAddrStr

	err = redis.SaveClientInfo(receiverRedisKey, ClientInfo)
	if err != nil {
		logger.Error("redis塞入错误:", zap.Error(err))
		return
	}

	// 预先分配 19 字节的切片
	replyToReceiver := make([]byte, 19)
	replyToReceiver[0] = 0x5A // 起始符 Z
	replyToReceiver[1] = 0x43 // 起始符 C
	replyToReceiver[2] = 0x14 // 命令码

	// 流水号
	copy(replyToReceiver[3:5], rawData[3:5])
	// 直接从收到的心跳包(rawData)中，零拷贝提取 8 字节的终端 ID
	copy(replyToReceiver[5:13], rawData[5:13])

	replyToReceiver[13] = 0x00 // 数据长度高位
	replyToReceiver[14] = 0x02 // 数据长度低位
	replyToReceiver[15] = 0x01 // 数据块 1
	replyToReceiver[16] = 0x01 // 数据块 2

	var checksum byte = 0
	for i := 2; i <= 16; i++ {
		checksum ^= replyToReceiver[i]
	}

	replyToReceiver[17] = checksum // 填入算好的动态校验码
	replyToReceiver[18] = 0x0D     // 结束符 \r

	// 替换掉 safeWriteUDP(server.conn, replyToReceiver, clientAddrStr)
	// 因为目标是 WS 发送端，所以用 WS 回传
	err = wsConn.WriteMessage(websocket.BinaryMessage, replyToReceiver)
	if err != nil {
		logger.Error("向接收机回复特定 0x14 指令失败:", zap.Error(err))
	}

	if ClientInfo.TransmitterHostPort != "" {
		serverAddr, err := net.ResolveUDPAddr("udp", ClientInfo.TransmitterHostPort) //发送
		if err != nil {
			log.Printf("连接 %s 失败：%v", err)
			return
		}
		if serverAddr.String() == clientAddrStr {
			log.Printf("🚨 严重网络串线拦截！APP地址与车端地址完全重叠，拒绝转发！", serverAddr.String())
			return
		}

		_, err = wsLocalUdpConn.WriteToUDP(rawData, serverAddr)
		if err != nil {
			log.Printf("回复发送端 %s 失败：%v", serverAddr.String(), err)
			return
		} else {
			return
		}
	}
}

// 发射机心跳
func handleWsTransmitterHeartBeat(wsConn *websocket.Conn, transmitterId string, rawData []byte, heartBeatPort string) {
	clientAddrStr := wsConn.RemoteAddr().String()

	receiverIdRedisKey := transmitterId //取到绑定的receiver
	receiverId := redis.Get(receiverIdRedisKey).Val()

	receiverRedisKey := string(receiverId) + "_receiver" //对应车辆配置信息 包含端口
	ClientInfo, err := redis.GetClientInfo(receiverRedisKey)
	if err != nil {
		logger.Error("未获取到redis缓存:", zap.Error(err))
		return
	}
	ClientInfo.TransmitterId = transmitterId

	clientAddrStrSet := clientAddrStr

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
	_, err = wsLocalUdpConn.WriteToUDP(rawData, serverAddr)
	if err != nil {
		log.Printf("发送消息到 %s 失败: %v", serverAddr.String(), err)
		return
	} else {
		return
	}
}

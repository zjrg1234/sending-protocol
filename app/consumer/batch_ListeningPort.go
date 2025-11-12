package consumer

import (
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"megin/library/logger"
	"net"
	"strings"
)

func startListeningPortReceiver(host string, port string) {
	listenAddr := host + ":" + port
	logger.Info("服务器开始监听端口" + listenAddr)

	//http.ListenAndServeTLS 预留上上ssl证书后使用该监听
	//err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil) https
	ips, err := net.LookupIP(host)
	fmt.Println(ips)

	ip := ips[0].String()
	if err != nil {
		logger.Info("域名解析失败：" + listenAddr)
		return
	}
	logger.Info("解析出ip" + ip + ":" + port)
	heartBeatPort := ":8898" //接收机接收心跳ip
	udpAddr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	fmt.Println("udpAddr", udpAddr)
	// 4. 创建 UDP 监听器
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Error("UDP 监听失败：", zap.Error(err))
		return
	}
	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("读取 UDP 数据失败: %v\n", err)
			continue
		}

		clientAddrStr := clientAddr.String() // 完整地址（IP:Port，如 "192.168.1.100:54321"）
		clientAddrHost := clientAddr.IP.String()
		//clientAddrPort := clientAddr.Port

		fmt.Printf("发送方ip加端口: %q\n", clientAddrStr) //测试使用后期注释

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
		commandCode := rawData[4:6]         //命令码
		// 4. 将十六进制字符串解码为字节切片

		if commandCode == "10" {
			go startForwardingReceiver(string(id), hexRawData, clientAddrStr, clientAddrHost, conn)
		}
		if commandCode == "15" {
			go getReceiverMessage(string(id), hexRawData, clientAddrStr)
		}
		if commandCode == "16" {
			go getReceiverHeartBeat(string(id), hexRawData, clientAddrStr, heartBeatPort)
		}
		fmt.Printf("获取到车辆id或发射机id: %q\n", id)
		fmt.Printf("命令码: %q\n", commandCode)

	}

}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ClientInfo struct {
	Addr     *net.UDPAddr
	ClientID string
	LastSeen time.Time
	Seq      int
}

type Server struct {
	conn    *net.UDPConn
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	port    int
}

func NewServer(port int) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		conn:    conn,
		clients: make(map[string]*ClientInfo),
		port:    port,
	}, nil
}

func (s *Server) handleMessage(data []byte, clientAddr *net.UDPAddr) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("解析客户端消息失败: %v", err)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientKey := clientAddr.String()

	// 更新或创建客户端信息
	if client, exists := s.clients[clientKey]; exists {
		client.LastSeen = time.Now()
		client.Seq = msg.Seq
	} else {
		s.clients[clientKey] = &ClientInfo{
			Addr:     clientAddr,
			ClientID: msg.ClientID,
			LastSeen: time.Now(),
			Seq:      msg.Seq,
		}
		log.Printf("新客户端注册: %s (%s)", msg.ClientID, clientKey)
	}

	switch msg.Type {
	case "register":
		log.Printf("客户端注册: %s", msg.ClientID)
		s.sendResponse(clientAddr, "welcome", "registration successful")

	case "heartbeat":
		log.Printf("心跳来自: %s", msg.ClientID)
		s.sendResponse(clientAddr, "heartbeat_ack", "heartbeat received")

	case "ack":
		log.Printf("收到ACK: %s, 序列号: %d", msg.ClientID, msg.Seq)

	default:
		log.Printf("未知消息类型: %s from %s", msg.Type, msg.ClientID)
	}
}

func (s *Server) sendResponse(addr *net.UDPAddr, msgType, data string) error {
	response := Message{
		Type:     msgType,
		Data:     data,
		Seq:      0,
		ClientID: "server",
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = s.conn.WriteToUDP(jsonData, addr)
	if err != nil {
		log.Printf("发送响应失败: %v", err)
		return err
	}

	log.Printf("发送响应到 %s: %s", addr.String(), msgType)
	return nil
}

func (s *Server) broadcastToClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.RLock()

		for clientKey, client := range s.clients {
			// 检查客户端是否活跃
			if time.Since(client.LastSeen) > 60*time.Second {
				log.Printf("客户端 %s 超时", client.ClientID)
				continue
			}

			// 发送测试数据
			msg := Message{
				Type:     "data",
				Data:     fmt.Sprintf("服务器时间: %v", time.Now()),
				Seq:      client.Seq + 1,
				ClientID: "server",
			}

			jsonData, err := json.Marshal(msg)
			if err != nil {
				log.Printf("序列化广播消息失败: %v", err)
				continue
			}

			_, err = s.conn.WriteToUDP(jsonData, client.Addr)
			if err != nil {
				log.Printf("广播到 %s 失败: %v", clientKey, err)
			} else {
				log.Printf("广播到 %s 成功", client.ClientID)
			}
		}

		s.mutex.RUnlock()
	}
}

func (s *Server) cleanupClients() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()

		for clientKey, client := range s.clients {
			if time.Since(client.LastSeen) > 120*time.Second {
				log.Printf("清理超时客户端: %s", client.ClientID)
				delete(s.clients, clientKey)
			}
		}

		s.mutex.Unlock()
	}
}

func (s *Server) Start() {
	log.Printf("UDP服务器启动在端口: %d", s.port)

	// 启动清理协程
	go s.cleanupClients()

	// 启动广播协程
	go s.broadcastToClients()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("读取数据失败: %v", err)
			continue
		}

		go s.handleMessage(buffer[:n], addr)
	}
}

func main() {
	server, err := NewServer(8899)
	if err != nil {
		log.Fatal(err)
	}

	server.Start()
}

package services

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"waf-backend/dto"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 프로덕션에서는 적절한 Origin 검증 필요
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketService struct {
	log        *logrus.Logger
	wafService *WAFService
	clients    map[*websocket.Conn]*Client
	clientsMux sync.RWMutex
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	conn     *websocket.Conn
	userID   string
	email    string
	send     chan []byte
}

type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewWebSocketService(log *logrus.Logger, wafService *WAFService) *WebSocketService {
	service := &WebSocketService{
		log:        log,
		wafService: wafService,
		clients:    make(map[*websocket.Conn]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	
	// WebSocket 허브 실행
	go service.run()
	
	// 주기적으로 WAF 통계 브로드캐스트
	go service.broadcastStats()
	
	return service
}

func (s *WebSocketService) HandleWebSocket(c *gin.Context) {
	// JWT에서 사용자 정보 추출 (이미 미들웨어에서 검증됨)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	email, _ := c.Get("email")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.WithError(err).Error("Failed to upgrade connection")
		return
	}
	
	client := &Client{
		conn:   conn,
		userID: userID.(string),
		email:  email.(string),
		send:   make(chan []byte, 256),
	}
	
	s.register <- client
	
	// 고루틴으로 클라이언트 처리
	go s.writePump(client)
	go s.readPump(client)
	
	s.log.WithFields(logrus.Fields{
		"user_id": client.userID,
		"email":   client.email,
	}).Info("WebSocket client connected")
}

func (s *WebSocketService) run() {
	for {
		select {
		case client := <-s.register:
			s.clientsMux.Lock()
			s.clients[client.conn] = client
			s.clientsMux.Unlock()
			
			// 새 클라이언트에게 환영 메시지와 초기 통계 전송
			welcome := WebSocketMessage{
				Type:      "welcome",
				Data:      gin.H{"message": "Connected to WAF Real-time Dashboard"},
				Timestamp: time.Now(),
			}
			
			if data, err := json.Marshal(welcome); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					s.clientsMux.Lock()
					delete(s.clients, client.conn)
					s.clientsMux.Unlock()
				}
			}
			
			// 초기 통계 전송
			stats := s.wafService.GetStats()
			statsMessage := WebSocketMessage{
				Type:      "stats",
				Data:      stats,
				Timestamp: time.Now(),
			}
			
			if data, err := json.Marshal(statsMessage); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					s.clientsMux.Lock()
					delete(s.clients, client.conn)
					s.clientsMux.Unlock()
				}
			}
			
		case client := <-s.unregister:
			s.clientsMux.Lock()
			if _, ok := s.clients[client.conn]; ok {
				delete(s.clients, client.conn)
				close(client.send)
				client.conn.Close()
			}
			s.clientsMux.Unlock()
			
		case message := <-s.broadcast:
			s.clientsMux.RLock()
			for conn, client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, conn)
				}
			}
			s.clientsMux.RUnlock()
		}
	}
}

func (s *WebSocketService) readPump(client *Client) {
	defer func() {
		s.unregister <- client
	}()
	
	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		var msg map[string]interface{}
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.WithError(err).Error("WebSocket error")
			}
			break
		}
		
		// 클라이언트로부터의 메시지 처리
		s.handleClientMessage(client, msg)
	}
}

func (s *WebSocketService) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *WebSocketService) handleClientMessage(client *Client, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}
	
	s.log.WithFields(logrus.Fields{
		"user_id": client.userID,
		"type":    msgType,
	}).Debug("Received WebSocket message")
	
	switch msgType {
	case "get_logs":
		// 최근 로그 요청
		limit := 50
		if limitVal, ok := msg["limit"].(float64); ok {
			limit = int(limitVal)
		}
		
		logs := s.wafService.GetLogs(limit)
		response := WebSocketMessage{
			Type:      "logs",
			Data:      logs,
			Timestamp: time.Now(),
		}
		
		if data, err := json.Marshal(response); err == nil {
			select {
			case client.send <- data:
			default:
				// 클라이언트 전송 버퍼가 가득 참
			}
		}
		
	case "get_stats":
		// 통계 요청
		stats := s.wafService.GetStats()
		response := WebSocketMessage{
			Type:      "stats",
			Data:      stats,
			Timestamp: time.Now(),
		}
		
		if data, err := json.Marshal(response); err == nil {
			select {
			case client.send <- data:
			default:
				// 클라이언트 전송 버퍼가 가득 참
			}
		}
	}
}

func (s *WebSocketService) broadcastStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		stats := s.wafService.GetStats()
		message := WebSocketMessage{
			Type:      "stats_update",
			Data:      stats,
			Timestamp: time.Now(),
		}
		
		if data, err := json.Marshal(message); err == nil {
			select {
			case s.broadcast <- data:
			default:
				// 브로드캐스트 채널이 가득 참
			}
		}
	}
}

func (s *WebSocketService) BroadcastNewLog(log *dto.WAFLog) {
	message := WebSocketMessage{
		Type:      "new_log",
		Data:      log,
		Timestamp: time.Now(),
	}
	
	if data, err := json.Marshal(message); err == nil {
		select {
		case s.broadcast <- data:
		default:
			// 브로드캐스트 채널이 가득 참
		}
	}
}

func (s *WebSocketService) GetConnectedClients() int {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	return len(s.clients)
}
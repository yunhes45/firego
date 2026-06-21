package transfer

import (
	"encoding/binary"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const CHUNK_SIZE = 64 * 1024 // 64KB

type Stream struct {
	sender    *websocket.Conn
	receiver  *websocket.Conn
	mu        sync.Mutex
	lastChunk uint64
	ready     chan struct{}
}

type StreamManager struct {
	streams map[string]*Stream
	mu      sync.RWMutex
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

func (sm *StreamManager) RegisterSender(sessionID string, conn *websocket.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[sessionID]; !exists {
		sm.streams[sessionID] = &Stream{
			ready: make(chan struct{}),
		}
	}
	sm.streams[sessionID].sender = conn
}

func (sm *StreamManager) RegisterReceiver(sessionID string, conn *websocket.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[sessionID]; !exists {
		sm.streams[sessionID] = &Stream{}
	}
	stream := sm.streams[sessionID]
	stream.receiver = conn
	close(stream.ready)
}

func (sm *StreamManager) Stream(sessionID string, session *Session) {
	sm.mu.RLock()
	stream, exists := sm.streams[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	timer := time.NewTimer(time.Until(session.ExpiresAt))
	defer timer.Stop()

	select {
	case <-stream.ready:
		stream.sender.WriteMessage(websocket.TextMessage, []byte("READY"))
	case <-timer.C:
		log.Println("TTL END", sessionID)
		sm.cleanup(sessionID)
		return
	}

	// A한테서 파일 받아서 청크로 쪼개서 B한테 전송
	pr, pw := io.Pipe()

	// A한테서 받기
	go func() {
		defer pw.Close()
		for {
			_, data, err := stream.sender.ReadMessage()
			if err != nil {
				log.Println("A 전송 종료:", err)
				return
			}
			pw.Write(data)
		}
	}()

	// 청크로 쪼개서 B한테 전송
	go func() {
		defer sm.cleanup(sessionID)
		buf := make([]byte, CHUNK_SIZE)
		var chunkIndex uint64 = 0

		for {
			n, err := pr.Read(buf)
			if n > 0 {
				// 청크 번호 + 데이터 합치기
				packet := make([]byte, 8+n)
				binary.BigEndian.PutUint64(packet[:8], chunkIndex)
				copy(packet[8:], buf[:n])

				stream.mu.Lock()
				werr := stream.receiver.WriteMessage(websocket.BinaryMessage, packet)
				stream.mu.Unlock()

				if werr != nil {
					log.Println("B 전송 오류:", werr)
					return
				}
				chunkIndex++
			}
			if err == io.EOF {
				log.Println("전송 완료:", sessionID)
				return
			}
			if err != nil {
				log.Println("청크 오류:", err)
				return
			}
		}
	}()
}

func (sm *StreamManager) cleanup(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[sessionID]; exists {
		if stream.sender != nil {
			stream.sender.Close()
		}
		if stream.receiver != nil {
			stream.receiver.Close()
		}
		delete(sm.streams, sessionID)
	}
}

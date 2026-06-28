package transfer

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const CHUNK_SIZE = 64 * 1024

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

func logf(streamID, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", streamID, msg)
}

func (sm *StreamManager) RegisterSender(streamID string, conn *websocket.Conn, filename string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[streamID]; !exists {
		sm.streams[streamID] = &Stream{
			ready: make(chan struct{}),
		}
	}
	sm.streams[streamID].sender = conn
	logf(streamID, "[%s] 송신자 접속", filename)
}

func (sm *StreamManager) RegisterReceiver(streamID string, conn *websocket.Conn, filename string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[streamID]; !exists {
		sm.streams[streamID] = &Stream{}
	}
	stream := sm.streams[streamID]
	stream.receiver = conn
	close(stream.ready)
	logf(streamID, "[%s] 수신자 접속", filename)
}

func (sm *StreamManager) Stream(streamID string, session *Session, filename string) {
	sm.mu.RLock()
	stream, exists := sm.streams[streamID]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	timer := time.NewTimer(time.Until(session.ExpiresAt))
	defer timer.Stop()

	select {
	case <-stream.ready:
		stream.sender.WriteMessage(websocket.TextMessage, []byte("READY"))
		logf(streamID, "[%s] 수신자 접속 확인 → READY 신호 전송", filename)
	case <-timer.C:
		logf(streamID, "[%s] TTL 만료 → 세션 종료", filename)
		sm.cleanup(streamID)
		return
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for {
			_, data, err := stream.sender.ReadMessage()
			if err != nil {
				logf(streamID, "[%s] 송신자 연결 종료: %v", filename, err)
				return
			}
			pw.Write(data)
		}
	}()

	go func() {
		defer sm.cleanup(streamID)
		buf := make([]byte, CHUNK_SIZE)
		var chunkIndex uint64 = 0
		var totalBytes uint64 = 0

		for {
			n, err := pr.Read(buf)
			if n > 0 {
				chunkIndex++
				totalBytes += uint64(n)
				logf(streamID, "[%s] 청크 %d 전송 (%d bytes)", filename, chunkIndex, n)

				stream.mu.Lock()
				werr := stream.receiver.WriteMessage(websocket.BinaryMessage, buf[:n])
				stream.mu.Unlock()

				if werr != nil {
					logf(streamID, "[%s] 수신자 전송 오류: %v", filename, werr)
					return
				}
			}
			if err == io.EOF {
				logf(streamID, "[%s] 전송 완료 (총 %d bytes, %d 청크)", filename, totalBytes, chunkIndex)
				return
			}
			if err != nil {
				logf(streamID, "[%s] 청크 처리 오류: %v", filename, err)
				return
			}
		}
	}()
}

func (sm *StreamManager) cleanup(streamID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[streamID]; exists {
		if stream.sender != nil {
			stream.sender.Close()
		}
		if stream.receiver != nil {
			stream.receiver.Close()
		}
		delete(sm.streams, streamID)
		logf(streamID, "세션 종료 및 정리 완료")
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"pfe/internal/transfer"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	sm  *transfer.SessionManager
	stm *transfer.StreamManager
}

func NewHandler(sm *transfer.SessionManager, stm *transfer.StreamManager) *Handler {
	return &Handler{sm: sm, stm: stm}
}

type CreateSessionRequest struct {
	Filename string `json:"filename"`
	TTL      string `json:"ttl"`
	Password string `json:"password"`
}

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	ExpiresAt string `json:"expires_at"`
}

// @Summary 세션 생성
// @Description 파일 전송 세션 생성
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "세션 정보"
// @Success 200 {object} CreateSessionResponse
// @Failure 400 {string} string "Invalid Request"
// @Router /session [post]
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	session, err := h.sm.CreateSession(req.Filename, req.TTL, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CreateSessionResponse{
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// @Summary 파일 전송 (A)
// @Description 웹소켓으로 파일 전송 시작
// @Param session_id path string true "세션 ID"
// @Router /send/{session_id} [get]
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/send/")

	session, exists := h.sm.GetSession(sessionID)
	if !exists {
		http.Error(w, "No Session", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket Connect Failed", http.StatusInternalServerError)
		return
	}

	h.stm.RegisterSender(sessionID, conn)
	go h.stm.Stream(sessionID, session)
}

// @Summary 파일 수신 (B)
// @Description 웹소켓으로 파일 수신
// @Param session_id path string true "세션 ID"
// @Router /receive/{session_id} [get]
func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/receive/")

	_, exists := h.sm.GetSession(sessionID)
	if !exists {
		http.Error(w, "No Session", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket Connect Failed", http.StatusInternalServerError)
		return
	}

	h.stm.RegisterReceiver(sessionID, conn)
}

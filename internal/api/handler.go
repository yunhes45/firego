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
	Files    []string `json:"files" example:"[\"a.zip\",\"b.zip\"]"`
	TTL      string   `json:"ttl" example:"1h"`
	Password string   `json:"password" example:"1234"`
}

type FileInfoResponse struct {
	FileID   string `json:"file_id"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

type CreateSessionResponse struct {
	GroupID   string             `json:"group_id"`
	Files     []FileInfoResponse `json:"files"`
	ExpiresAt string             `json:"expires_at"`
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

	session, err := h.sm.CreateSession(req.Files, req.TTL, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := make([]FileInfoResponse, len(session.Files))
	for i, f := range session.Files {
		files[i] = FileInfoResponse{
			FileID:   f.FileID,
			Filename: f.Filename,
			Status:   f.Status,
		}
	}

	resp := CreateSessionResponse{
		GroupID:   session.GroupID,
		Files:     files,
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
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/send/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}

	groupID := parts[0]
	fileID := parts[1]

	session, exists := h.sm.GetSession(groupID)
	if !exists {
		http.Error(w, "No Session", http.StatusNotFound)
		return
	}

	var filename string
	for _, f := range session.Files {
		if f.FileID == fileID {
			filename = f.Filename
			break
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket Connect Failed", http.StatusInternalServerError)
		return
	}

	h.stm.RegisterSender(groupID+"/"+fileID, conn, filename)
	go h.stm.Stream(groupID+"/"+fileID, session, filename)
}

// @Summary 파일 수신 (B)
// @Description 웹소켓으로 파일 수신
// @Param session_id path string true "세션 ID"
// @Router /receive/{session_id} [get]
func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/receive/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}
	groupID := parts[0]
	fileID := parts[1]

	session, exists := h.sm.GetSession(groupID)
	if !exists {
		http.Error(w, "No Session", http.StatusNotFound)
		return
	}

	var filename string
	for _, f := range session.Files {
		if f.FileID == fileID {
			filename = f.Filename
			break
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket Connect Failed", http.StatusInternalServerError)
		return
	}

	h.stm.RegisterReceiver(groupID+"/"+fileID, conn, filename)
}

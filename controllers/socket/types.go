package socket

import (
	socketio "github.com/googollee/go-socket.io"
	"team.gg-server/models"
)

// types
const (
	EventTest = "test"
)

type UserSocket struct {
	User *models.UserDAO
	Conn socketio.Conn
}

type Manager struct {
	users   map[string]UserSocket
	sockets map[string]UserSocket
	Io      *socketio.Server
}

func NewSocketManager() *Manager {
	return &Manager{
		users:   make(map[string]UserSocket),
		sockets: make(map[string]UserSocket),
	}
}

func (sm *Manager) AddUser(user *models.UserDAO, conn socketio.Conn) {
	userSocket := UserSocket{
		Conn: conn,
	}
	if user != nil {
		userSocket.User = user
		sm.users[user.UserId] = userSocket
	}
	sm.sockets[conn.ID()] = userSocket
}

func (sm *Manager) RemoveUserByUserId(userId string) {
	userSocket, ok := sm.users[userId]
	if ok {
		delete(sm.sockets, userSocket.Conn.ID())
		delete(sm.users, userId)
	}
}

func (sm *Manager) RemoveUserByConnId(connId string) {
	userSocket, ok := sm.sockets[connId]
	if ok {
		delete(sm.sockets, connId)
		delete(sm.users, userSocket.User.UserId)
	}
}

func (sm *Manager) GetUserByUserId(userId string) (UserSocket, bool) {
	userSocket, ok := sm.users[userId]
	return userSocket, ok
}

// -------------------------------------------------------------------------------------

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *string     `json:"error"`
}

func NewResponse(success bool, data interface{}, err *string) Response {
	return Response{
		Success: success,
		Data:    data,
		Error:   err,
	}
}

func NewSuccess(data interface{}) Response {
	return NewResponse(true, data, nil)
}

func NewFailure(errMsg string) Response {
	return NewResponse(false, nil, &errMsg)
}

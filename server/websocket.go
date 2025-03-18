package server

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

//go:generate mockgen -destination=./mock/websocket.go -package mock github.com/xescugc/maze-wars/server WSConnector
type WSConnector interface {
	Write(context.Context, *websocket.Conn, interface{}) error
	Read(context.Context, *websocket.Conn, interface{}) error
}

type WS struct{}

func NewWS() *WS { return &WS{} }

func (ws *WS) Write(ctx context.Context, conn *websocket.Conn, d interface{}) error {
	return wsjson.Write(ctx, conn, d)
}

func (ws *WS) Read(ctx context.Context, conn *websocket.Conn, d interface{}) error {
	return wsjson.Read(ctx, conn, d)
}

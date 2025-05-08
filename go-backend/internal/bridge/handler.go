package bridge

import (
	"context"
	"encoding/json"
	"go-backend/internal/power"
	"io"
	"log"
	"net"
)

type Request struct {
	Action string `json:"action"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req Request
	if err := decoder.Decode(&req); err != nil {
		log.Printf("❌ Invalid request: %v", err)
		_ = encoder.Encode(Response{Status: "error", Error: "invalid JSON"})
		return
	}

	var res Response
	switch req.Action {
	case "reboot":
		err := power.RebootSystem(context.Background())
		if err != nil {
			res = Response{Status: "error", Error: err.Error()}
		} else {
			res = Response{Status: "ok", Message: "rebooting"}
		}
	case "shutdown":
		err := power.PowerOffSystem(context.Background())
		if err != nil {
			res = Response{Status: "error", Error: err.Error()}
		} else {
			res = Response{Status: "ok", Message: "shutting down"}
		}
	default:
		res = Response{Status: "error", Error: "unknown action"}
	}

	if err := encoder.Encode(res); err != nil && err != io.EOF {
		log.Printf("❌ Failed to write response: %v", err)
	}
}

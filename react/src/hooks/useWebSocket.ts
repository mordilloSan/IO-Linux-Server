// useWebSocket.ts
import { useEffect, useState, useCallback } from "react";

export function useWebSocket(enabled: boolean) {
  const [socket, setSocket] = useState<WebSocket | null>(null);

  const disconnect = useCallback(() => {
    if (socket) {
      console.log("🔌 Disconnecting WebSocket manually");
      socket.close();
      setSocket(null);
    }
  }, [socket]);

  useEffect(() => {
    if (!enabled) {
      disconnect();
      return;
    }

    const ws = new WebSocket("/ws/system");
    setSocket(ws);

    ws.onopen = () => console.log("✅ WebSocket connected");
    ws.onclose = () => {
      console.log("❌ WebSocket disconnected");
      setSocket(null);
    };
    ws.onerror = (err) => console.error("WebSocket error:", err);

    return () => {
      ws.close();
      setSocket(null);
    };
  }, [enabled, disconnect]);

  return { socket, disconnect };
}

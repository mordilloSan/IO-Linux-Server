import React, { createContext, useRef, useEffect } from "react";

type MessageHandler = (data: any) => void;

interface WebSocketContextValue {
  subscribe: (channel: string, handler: MessageHandler) => void;
  unsubscribe: () => void;
}

export const WebSocketContext = createContext<WebSocketContextValue | null>(
  null,
);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const wsRef = useRef<WebSocket | null>(null);
  const channelRef = useRef<string | null>(null);
  const handlerRef = useRef<MessageHandler | null>(null);

  useEffect(() => {
    const ws = new WebSocket("/ws");
    console.log("🧠 WS connecting to /ws");
    wsRef.current = ws;

    ws.onopen = () => {
      if (channelRef.current) {
        ws.send(
          JSON.stringify({ type: "subscribe", channel: channelRef.current }),
        );
      }
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("📨 WS message received:", data); // ✅ Add this
        handlerRef.current?.(data);
      } catch (e) {
        console.error("Invalid WebSocket message", e);
      }
    };

    ws.onerror = (e) => console.error("WebSocket error:", e);
    ws.onclose = () => console.log("WebSocket closed");

    return () => {
      ws.close();
    };
  }, []);

  const subscribe = (channel: string, handler: MessageHandler) => {
    channelRef.current = channel;
    handlerRef.current = handler;

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "subscribe", channel }));
    }
  };

  const unsubscribe = () => {
    channelRef.current = null;
    handlerRef.current = null;
  };

  return (
    <WebSocketContext.Provider value={{ subscribe, unsubscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

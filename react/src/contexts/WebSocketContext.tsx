import React, {
  createContext,
  useRef,
  useEffect,
  useCallback,
  useState,
} from "react";

type Message = { type: string; channel: string; payload: any };
type MessageHandler = (payload: any) => void;

interface WebSocketContextValue {
  subscribe: (channel: string, handler: MessageHandler) => void;
  unsubscribe: (channel: string, handler: MessageHandler) => void;
  latestData: Record<string, any>;
}

export const WebSocketContext = createContext<WebSocketContextValue | null>(
  null
);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const wsRef = useRef<WebSocket | null>(null);
  const listeners = useRef<Map<string, Set<MessageHandler>>>(new Map());
  const [latestData, setLatestData] = useState<Record<string, any>>({});

  useEffect(() => {
    const wsUrl = import.meta.env.VITE_API_WS_URL || "ws://localhost:8080/ws";

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("✅ WebSocket connected");

      for (const channel of listeners.current.keys()) {
        console.log(`📨 Subscribing to '${channel}'`);
        ws.send(JSON.stringify({ action: "subscribe", channel }));
      }
    };

    ws.onmessage = (event) => {
      try {
        // Uncomment for debugging
        // console.log("📨 Raw message:", event.data);

        const msg: Message = JSON.parse(event.data);
        const { type, channel, payload } = msg;

        // Uncomment for debugging
        //console.log("✅ Parsed message:", { type, channel, payload });

        setLatestData((prev) => ({ ...prev, [type]: payload }));

        const handlers = listeners.current.get(channel);
        if (handlers) {
          handlers.forEach((h) => h(payload));
        }
      } catch (e) {
        console.error("❌ Failed to parse WebSocket message", e);
      }
    };

    ws.onclose = () => console.log("🔌 WebSocket disconnected");
    ws.onerror = (e) => console.error("⚠️ WebSocket error", e);

    return () => ws.close();
  }, []);

  const subscribe = useCallback((channel: string, handler: MessageHandler) => {
    let channelHandlers = listeners.current.get(channel);
    if (!channelHandlers) {
      channelHandlers = new Set();
      listeners.current.set(channel, channelHandlers);
    }
    channelHandlers.add(handler);

    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ action: "subscribe", channel }));
    } else {
      // queue for later send in ws.onopen
      console.log(`📨 Queued subscription to '${channel}'`);
    }
  }, []);

  const unsubscribe = useCallback(
    (channel: string, handler: MessageHandler) => {
      const handlers = listeners.current.get(channel);
      if (!handlers) return;
      handlers.delete(handler);
      if (handlers.size === 0) {
        listeners.current.delete(channel);
        wsRef.current?.send(JSON.stringify({ action: "unsubscribe", channel }));
      }
    },
    []
  );

  return (
    <WebSocketContext.Provider value={{ subscribe, unsubscribe, latestData }}>
      {children}
    </WebSocketContext.Provider>
  );
};

import React, {
  createContext,
  useRef,
  useEffect,
  useCallback,
  useState,
  useContext,
} from "react";
import { AuthContext } from "./AuthContext";

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
  const reconnectTimeout = useRef<number | null>(null);
  const retryCount = useRef(0);

  const listeners = useRef<Map<string, Set<MessageHandler>>>(new Map());
  const [latestData, setLatestData] = useState<Record<string, any>>({});

  const connect = useCallback(() => {
    // Prevent connecting if already connected or connecting
    if (
      wsRef.current &&
      (wsRef.current.readyState === WebSocket.OPEN ||
        wsRef.current.readyState === WebSocket.CONNECTING)
    ) {
      console.log("🔁 WebSocket already connected or connecting");
      return;
    }

    const ws = new WebSocket("/ws");
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
      retryCount.current = 0;

      for (const channel of listeners.current.keys()) {
        console.log(`📨 Re-subscribing to '${channel}'`);
        ws.send(JSON.stringify({ action: "subscribe", channel }));
      }
    };

    ws.onmessage = (event) => {
      try {
        const msg: Message = JSON.parse(event.data);
        const { type, channel, payload } = msg;

        setLatestData((prev) => ({ ...prev, [type]: payload }));

        const handlers = listeners.current.get(channel);
        if (handlers) {
          handlers.forEach((h) => h(payload));
        }
      } catch (e) {
        console.error("❌ Failed to parse WebSocket message", e);
      }
    };

    ws.onclose = () => {
      console.log("🔌 WebSocket disconnected");
      wsRef.current = null;
      reconnect();
    };

    ws.onerror = (e) => {
      console.error("⚠️ WebSocket error", e);
      ws.close(); // Will trigger onclose
    };
  }, []);

  const reconnect = useCallback(() => {
    if (reconnectTimeout.current) return;

    const delay = Math.min(1000 * 2 ** retryCount.current, 10000);
    console.log(`⏳ Reconnecting in ${delay / 1000}s...`);

    reconnectTimeout.current = setTimeout(() => {
      retryCount.current++;
      reconnectTimeout.current = null;
      connect();
    }, delay);
  }, [connect]);

  const auth = useContext(AuthContext);

  useEffect(() => {
    if (auth?.isAuthenticated) {
      connect();
    } else {
      wsRef.current?.close();
    }

    return () => {
      reconnectTimeout.current && clearTimeout(reconnectTimeout.current);

      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.close();
      }
    };
  }, [auth?.isAuthenticated, connect]);

  const subscribe = useCallback((channel: string, handler: MessageHandler) => {
    let handlers = listeners.current.get(channel);
    if (!handlers) {
      handlers = new Set();
      listeners.current.set(channel, handlers);

      if (wsRef.current?.readyState === WebSocket.OPEN) {
        console.log(`📨 Subscribing to new channel '${channel}'`);
        wsRef.current.send(JSON.stringify({ action: "subscribe", channel }));
      } else {
        console.log(`🕓 Delaying subscription to '${channel}' until reconnect`);
      }
    }

    handlers.add(handler);
  }, []);

  const unsubscribe = useCallback(
    (channel: string, handler: MessageHandler) => {
      const handlers = listeners.current.get(channel);
      if (!handlers) return;

      handlers.delete(handler);

      if (handlers.size === 0) {
        listeners.current.delete(channel);
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          console.log(`📭 Unsubscribing from channel '${channel}'`);
          wsRef.current.send(
            JSON.stringify({ action: "unsubscribe", channel })
          );
        }
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

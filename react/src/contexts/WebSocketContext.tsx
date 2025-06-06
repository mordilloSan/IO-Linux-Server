import React, {
  createContext,
  useContext,
  useRef,
  useEffect,
  useState,
  useCallback,
} from "react";

import useAuth from "@/hooks/useAuth";

type WebSocketContextValue = {
  ws: WebSocket | null;
  send: (msg: any) => void;
  lastMessage: any;
};

const WebSocketContext = createContext<WebSocketContextValue>({
  ws: null,
  send: () => {},
  lastMessage: null,
});

export const useWebSocket = () => useContext(WebSocketContext);

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [lastMessage, setLastMessage] = useState<any>(null);
  const ws = useRef<WebSocket | null>(null);
  const { isAuthenticated, isInitialized } = useAuth();

  useEffect(() => {
    if (!isInitialized || !isAuthenticated) return;

    const wsUrl = import.meta.env.DEV
      ? "ws://localhost:8080/ws"
      : window.location.protocol === "https:"
        ? `wss://${window.location.host}/ws`
        : `ws://${window.location.host}/ws`;

    const socket = new WebSocket(wsUrl);
    ws.current = socket;

    socket.onmessage = (event) => {
      try {
        setLastMessage(JSON.parse(event.data));
      } catch {
        setLastMessage(event.data);
      }
    };

    socket.onerror = (event) => {
      console.error("WebSocket error", event);
    };

    socket.onopen = () => {
      console.log("WebSocket opened!");
    };

    socket.onclose = () => {
      console.log("WebSocket closed!");
    };

    return () => {
      socket.close();
    };
  }, [isAuthenticated, isInitialized]);

  // ---- Memoize send! ----
  const send = useCallback((msg: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(msg));
    }
  }, []);

  if (!isInitialized || !isAuthenticated) return null;

  return (
    <WebSocketContext.Provider value={{ ws: ws.current, send, lastMessage }}>
      {children}
    </WebSocketContext.Provider>
  );
};

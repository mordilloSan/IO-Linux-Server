import React, {
  createContext,
  useEffect,
  useRef,
  useState,
  ReactNode,
} from "react";
import useAuth from "@/hooks/useAuth";

type WebSocketValue = {
  socket: WebSocket | null;
  disconnect: () => void;
};

export const WebSocketContext = createContext<WebSocketValue>({
  socket: null,
  disconnect: () => {},
});

export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const { isAuthenticated } = useAuth();
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const socketRef = useRef<WebSocket | null>(null);

  const disconnect = () => {
    if (socketRef.current) {
      console.log("🔌 Disconnecting WebSocket manually");
      socketRef.current.close();
      socketRef.current = null;
      setSocket(null);
    }
  };

  useEffect(() => {
    if (!isAuthenticated) {
      disconnect();
      return;
    }

    const ws = new WebSocket("ws://localhost:8080/ws/system");
    socketRef.current = ws;
    setSocket(ws);

    ws.onopen = () => console.log("✅ WebSocket connected");
    ws.onclose = () => {
      console.log("❌ WebSocket disconnected");
      socketRef.current = null;
      setSocket(null);
    };
    ws.onerror = (err) => console.error("WebSocket error:", err);

    return () => {
      ws.close();
    };
  }, [isAuthenticated]);

  return (
    <WebSocketContext.Provider value={{ socket, disconnect }}>
      {children}
    </WebSocketContext.Provider>
  );
};

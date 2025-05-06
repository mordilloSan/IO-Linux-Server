import { createContext, ReactNode } from "react";

import useAuth from "@/hooks/useAuth";
import { useWebSocket } from "@/hooks/useWebSocket";

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
  const { socket, disconnect } = useWebSocket(isAuthenticated);

  return (
    <WebSocketContext.Provider value={{ socket, disconnect }}>
      {children}
    </WebSocketContext.Provider>
  );
};

import { useContext } from "react";
import { WebSocketContext } from "@/contexts/WebSocketContext";

export const useWebSocket = () => {
  const ctx = useContext(WebSocketContext);
  if (!ctx) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return ctx;
};

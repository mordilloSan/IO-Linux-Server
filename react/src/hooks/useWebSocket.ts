import { WebSocketContext } from "@/contexts/WebSocketContext";
import { useContext } from "react";

export function useWebSocket() {
  const ctx = useContext(WebSocketContext);
  if (!ctx) {
    throw new Error("useWebSocket must be used within WebSocketProvider");
  }
  return ctx;
}

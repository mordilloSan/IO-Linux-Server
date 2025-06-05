import { useWebSocket } from "../contexts/WebSocketContext";
import { useEffect, useState } from "react";

// Utility for unique IDs per request
function generateRequestId() {
  return Math.random().toString(36).slice(2) + Date.now();
}

// WebSocket query function for React Query
export function useWSQuery<TData = any>(
  type: string,
  payload: any,
  options?: { enabled?: boolean },
) {
  const { ws } = useWebSocket();
  const [data, setData] = useState<TData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!ws || ws.readyState !== WebSocket.OPEN || options?.enabled === false) {
      return;
    }

    setLoading(true);
    setData(null);
    setError(null);

    const requestId = generateRequestId();

    const handleMessage = (event: MessageEvent) => {
      let message;
      try {
        message = JSON.parse(event.data);
      } catch {
        return;
      }
      if (message.requestId !== requestId) return;

      if (message.error) setError(message.error);
      else setData(message.data);

      setLoading(false);
      ws.removeEventListener("message", handleMessage);
    };

    ws.addEventListener("message", handleMessage);

    ws.send(
      JSON.stringify({
        type,
        requestId,
        payload,
      }),
    );

    // Cleanup
    return () => ws.removeEventListener("message", handleMessage);
  }, [type, JSON.stringify(payload), ws, options?.enabled]);

  return { data, error, loading };
}

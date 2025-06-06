import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { useWebSocket } from "@/contexts/WebSocketContext";
import { ROUTE_CHANNELS } from "../routes";

// Returns channel for current location
function getChannelForPath(pathname: string) {
  const clean = pathname.replace(/^\//, "").split("/")[0];
  // Map "" to "dashboard" (the root)
  if (!clean) return "dashboard";
  // Only auto-subscribe if in known list
  return ROUTE_CHANNELS.includes(clean) ? clean : null;
}

export function useRouteChannelSubscription() {
  const { send } = useWebSocket();
  const location = useLocation();
  const channel = getChannelForPath(location.pathname);

  useEffect(() => {
    if (!channel) {
      console.log(
        `[WS Channel] No matching channel for pathname "${location.pathname}"`,
      );
      return;
    }
    send({ type: "subscribe", payload: { channel } });

    return () => {
      send({ type: "unsubscribe", payload: { channel } });
    };
  }, [channel, send, location.pathname]);
}

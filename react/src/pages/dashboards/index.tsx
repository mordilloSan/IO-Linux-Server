import React, { useContext, useEffect, useState } from "react";
import {
  Grid,
  Typography,
  Paper as MuiPaper,
  Divider as MuiDivider,
  Box,
} from "@mui/material";
import { spacing } from "@mui/system";
import styled from "@emotion/styled";
import { WebSocketContext } from "@/contexts/WebSocketContext";

const Paper = styled(MuiPaper)(spacing);
const Divider = styled(MuiDivider)(spacing);

type SystemStats = {
  cpu: number[];
  memory: {
    total: number;
    used: number;
    free: number;
    usedPercent: number;
    [key: string]: any;
  };
  network: {
    [iface: string]: {
      bytesSent: number;
      bytesRecv: number;
    };
  };
};

const formatBytes = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`;
  const kb = bytes / 1024;
  if (kb < 1024) return `${kb.toFixed(1)} KB`;
  const mb = kb / 1024;
  return `${mb.toFixed(1)} MB`;
};

const LiveSystemStats: React.FC = () => {
  const { socket } = useContext(WebSocketContext);
  const [stats, setStats] = useState<SystemStats | null>(null);

  useEffect(() => {
    if (!socket) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        setStats(data);
      } catch (err) {
        console.error("❌ Failed to parse WebSocket data:", err);
      }
    };

    socket.addEventListener("message", handleMessage);
    return () => {
      socket.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <Grid container spacing={6}>
      <Grid size={12}>
        <Typography variant="h3" gutterBottom>
          Live System Stats
        </Typography>
        <Divider my={4} />
      </Grid>

      <Grid size={12}>
        <Paper p={4}>
          {!stats ? (
            <Typography>Loading system stats...</Typography>
          ) : (
            <>
              <Typography variant="h6">🧠 CPU</Typography>
              <Box mb={2}>
                {stats.cpu.map((core, i) => (
                  <Typography key={i}>
                    Core {i + 1}: {core.toFixed(1)}%
                  </Typography>
                ))}
              </Box>

              <Typography variant="h6">💾 Memory</Typography>
              <Box mb={2}>
                <Typography>
                  {stats.memory.usedPercent.toFixed(1)}% used (
                  {(stats.memory.used / 1024 / 1024 / 1024).toFixed(1)} GB of{" "}
                  {(stats.memory.total / 1024 / 1024 / 1024).toFixed(1)} GB)
                </Typography>
              </Box>

              <Typography variant="h6">🌐 Network</Typography>
              <Box>
                {Object.entries(stats.network).map(([iface, net]) => (
                  <Typography key={iface}>
                    {iface}: ↑ {formatBytes(net.bytesSent)} ↓{" "}
                    {formatBytes(net.bytesRecv)}
                  </Typography>
                ))}
              </Box>
            </>
          )}
        </Paper>
      </Grid>
    </Grid>
  );
};

export default LiveSystemStats;

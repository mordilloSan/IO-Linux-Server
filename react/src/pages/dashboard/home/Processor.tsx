import React from "react";
import { Box, Typography } from "@mui/material";
import TemperatureIcon from "@mui/icons-material/Thermostat";

import GeneralCard from "@/components/cards/GeneralCard";
import CircularProgressWithLabel from "@/components/gauge/CircularProgress";
import ComponentLoader from "@/components/loaders/ComponentLoader";
import { useWebSocket } from "@/hooks/useWebSocket";

interface CPUInfoResponse {
  vendorId: string;
  modelName: string;
  family: string;
  model: string;
  mhz: number;
  cores: number;
  loadAverage: {
    load1: number;
    load5: number;
    load15: number;
  };
  perCoreUsage: number[];
  temperature: { [core: string]: number };
}

const Processor: React.FC = () => {
  const { latestData } = useWebSocket();
  const CPUInfo = latestData["cpu"] as CPUInfoResponse | undefined;

  if (!CPUInfo) return <ComponentLoader />;

  const averageCpuUsage = CPUInfo.perCoreUsage?.length
    ? CPUInfo.perCoreUsage.reduce((sum: number, cpu: number) => sum + cpu, 0) /
      CPUInfo.perCoreUsage.length
    : 0;

  const temperatures = Object.values(CPUInfo.temperature || {});
  const avgTemp = temperatures.length
    ? (
        temperatures.reduce((sum: number, t: number) => sum + t, 0) /
        temperatures.length
      ).toFixed(1)
    : "--";

  const IconText = `${avgTemp}°C`;

  const data = {
    title: "Processor",
    avatarIcon: "ph:cpu",
    stats2: (
      <CircularProgressWithLabel
        value={averageCpuUsage}
        size={120}
        thickness={4}
      />
    ),
    stats: (
      <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
        <Typography variant="body1">
          <strong>CPU:</strong> {CPUInfo.modelName}
        </Typography>
        <Typography variant="body1">
          <strong>Cores:</strong> {CPUInfo.cores} Threads
        </Typography>
        <Typography variant="body1">
          <strong>Max Usage:</strong>{" "}
          {Math.max(...(CPUInfo.perCoreUsage || [0])).toFixed(0)}%
        </Typography>
      </Box>
    ),
    icon_text: IconText,
    icon: TemperatureIcon,
    iconProps: { sx: { color: "grey" } },
  };

  return <GeneralCard {...data} />;
};

export default Processor;

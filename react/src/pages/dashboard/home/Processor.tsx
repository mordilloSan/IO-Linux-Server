"use client";

import React from "react";
import TemperatureIcon from "@mui/icons-material/Thermostat";
import { Box, Typography } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import CardWithBorder from "@/components/cards/CardWithBorder";
import CircularProgressWithLabel from "@/components/CircularProgress";
import axios from "@/utils/axios";
import ErrorMessage from "@/components/Error";
import ComponentLoader from "@/components/ComponentLoader";

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
  const {
    data: CPUInfo,
    isPending,
    isError,
  } = useQuery<CPUInfoResponse>({
    queryKey: ["CPUInfo"],
    queryFn: async () => {
      const response = await axios.get("/system/cpu");
      return response.data;
    },
    refetchInterval: 2000,
  });

  const averageCpuUsage = CPUInfo?.perCoreUsage
    ? CPUInfo.perCoreUsage.reduce((sum, cpu) => sum + cpu, 0) /
      CPUInfo.perCoreUsage.length
    : 0;

  const temperatures = CPUInfo?.temperature
    ? Object.values(CPUInfo.temperature)
    : [];

  const avgTemp = temperatures.length
    ? (
        temperatures.reduce((sum, t) => sum + t, 0) / temperatures.length
      ).toFixed(1)
    : "--";

  const IconText = `${avgTemp}°C`;

  const data = {
    title: "Processor",
    avatarIcon: "ph:cpu",
    stats: isError ? (
      <ErrorMessage />
    ) : isPending ? (
      <ComponentLoader />
    ) : (
      <CircularProgressWithLabel
        value={averageCpuUsage}
        size={120}
        thickness={4}
      />
    ),
    stats2: (
      <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
        <Typography variant="body1">
          <strong>CPU:</strong> {CPUInfo?.modelName}
        </Typography>
        <Typography variant="body1">
          <strong>Cores:</strong> {CPUInfo?.cores} Threads
        </Typography>
        <Typography variant="body1">
          <strong>Max Usage:</strong>{" "}
          {Math.max(...(CPUInfo?.perCoreUsage || [0])).toFixed(0)}%
        </Typography>
      </Box>
    ),
    icon_text: IconText,
    icon: TemperatureIcon,
    iconProps: { sx: { color: "grey" } },
  };

  return <CardWithBorder {...data} />;
};

export default Processor;

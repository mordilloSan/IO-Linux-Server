import React from "react";
import { useQuery } from "@tanstack/react-query";
import { Typography, Box } from "@mui/material";

import axios from "@/utils/axios"; // ✅ use your axios instance
import CardWithBorder from "@/components/cards/CardWithBorder";
import CircularProgressWithLabel from "@/components/CircularProgress";
import ErrorMessage from "@/components/Error";
import { Loader } from "lucide-react";

// Utility functions
const formatBytesToGB = (bytes: number) => (bytes / 1000 ** 3).toFixed(2);
const calculatePercentage = (used: number, total: number) =>
  ((used / total) * 100).toFixed(2);

const MemoryUsage = () => {
  const {
    data: memoryData,
    isPending,
    isError,
  } = useQuery({
    queryKey: ["memoryInfo"],
    queryFn: async () => {
      const response = await axios.get("/system/mem");
      return response.data;
    },
    refetchInterval: 2000,
  });

  const ramUsagePercentage = memoryData?.active
    ? parseFloat(calculatePercentage(memoryData.active, memoryData.total))
    : 0;

  const data = {
    title: "Memory Usage",
    stats2: isError ? (
      <ErrorMessage />
    ) : isPending ? (
      <Loader />
    ) : (
      <CircularProgressWithLabel
        value={ramUsagePercentage}
        size={120}
        thickness={4}
      />
    ),
    stats: (
      <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
        <Typography variant="body1">
          <strong>Total Memory:</strong>{" "}
          {formatBytesToGB(memoryData?.total || 0)} GB
        </Typography>
        <Typography variant="body1">
          <strong>Used Memory:</strong>{" "}
          {formatBytesToGB(memoryData?.active || 0)} GB
        </Typography>
        <Typography variant="body1">
          <strong>Swap:</strong>{" "}
          {formatBytesToGB(memoryData?.swapTotal - memoryData?.swapFree || 0)}{" "}
          of {formatBytesToGB(memoryData?.swapTotal || 0)} GB
        </Typography>
      </Box>
    ),
    avatarIcon: "la:memory",
  };

  return <CardWithBorder {...data} />;
};

export default MemoryUsage;

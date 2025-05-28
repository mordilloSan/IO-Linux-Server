import React from "react";
import { Box, Typography, CircularProgress, Alert } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import axios from "@/utils/axios";
import ServiceTable, { Service } from "./ServiceTable";

const fetchServices = async (): Promise<Service[]> => {
  const res = await axios.get("/system/services/status");
  return res.data;
};

const ServicesList: React.FC = () => {
  // 5s refetch for near-realtime, adjust as needed
  const { data, isLoading, isError, error, refetch } = useQuery<Service[]>({
    queryKey: ["services"],
    queryFn: fetchServices,
    refetchInterval: 5000,
  });

  // Handlers - you would implement actual API calls here!
  const handleRestart = (service: Service) => {
    // Example: await axios.post(`/services/restart`, { name: service.name })
    alert(`Restarting service: ${service.displayName || service.name}`);
    // After API: refetch();
  };

  const handleStop = (service: Service) => {
    alert(`Stopping service: ${service.displayName || service.name}`);
    // After API: refetch();
  };

  const handleViewLogs = (service: Service) => {
    alert(`Show logs for: ${service.displayName || service.name}`);
    // You might open a drawer/modal with logs, etc.
  };

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Services
      </Typography>
      {isLoading && (
        <Box textAlign="center" my={5}>
          <CircularProgress />
        </Box>
      )}
      {isError && (
        <Alert severity="error">
          {error instanceof Error ? error.message : "Failed to load services"}
        </Alert>
      )}
      {data && (
        <ServiceTable
          serviceList={data}
          onRestart={handleRestart}
          onStop={handleStop}
          onViewLogs={handleViewLogs}
        />
      )}
    </Box>
  );
};

export default ServicesList;

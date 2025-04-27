import React from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Box, Typography, Grid } from "@mui/material";
import ContainerCard from "./ContainerCard";
import { ContainerInfo } from "@/types/container";
import axios from "@/utils/axios";

const ContainerList: React.FC = () => {
  const queryClient = useQueryClient();

  const { data: containers = [], isLoading } = useQuery<ContainerInfo[]>({
    queryKey: ["containers"],
    queryFn: async () => {
      const res = await axios.get("/docker/containers");
      return res.data;
    },
    refetchInterval: 5000,
  });

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 2 }}>
        Containers
      </Typography>

      <Grid container spacing={2}>
        {containers.map((container) => (
          <ContainerCard
            key={container.Id}
            container={container}
            queryClient={queryClient}
          />
        ))}
      </Grid>

      {isLoading && (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
          Loading containers...
        </Typography>
      )}
    </Box>
  );
};

export default ContainerList;

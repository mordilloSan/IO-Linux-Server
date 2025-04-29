import React, { Suspense } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Box, Typography, Grid } from "@mui/material";
import ContainerCard from "./ContainerCard";
import { ContainerInfo } from "@/types/container";
import axios from "@/utils/axios";

const ContainerList: React.FC = () => {
  const queryClient = useQueryClient();

  const { data: containers = [] } = useQuery<ContainerInfo[]>({
    queryKey: ["containers"],
    queryFn: async () => {
      const res = await axios.get("/docker/containers");
      return res.data;
    },
    refetchInterval: 5000,
  });

  return (
    <Suspense fallback={<Typography>Loading containers...</Typography>}>
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
      </Box>
    </Suspense>
  );
};

export default ContainerList;

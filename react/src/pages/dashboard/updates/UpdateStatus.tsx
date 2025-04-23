// components/system/UpdateStatus.tsx
import React from "react";
import { Box, Typography } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import axios from "@/utils/axios";
import UpdateActions from "./UpdateActions";
import UpdateList from "./UpdateList";
import { Update } from "@/types/update";

interface UpdateInfo {
  updates: Update[];
}

const UpdateStatus: React.FC = () => {
  const {
    data: updateInfo,
    isLoading,
    refetch,
  } = useQuery<UpdateInfo>({
    queryKey: ["updateInfo"],
    queryFn: () => axios.get("/system/updates").then((res) => res.data),
    refetchInterval: 50000,
  });

  const updates = updateInfo?.updates || [];

  const renderCollapseContent = (row: Update) => (
    <Box sx={{ whiteSpace: "pre-wrap", fontSize: 14 }}>
      <Typography variant="body2" color="text.secondary">
        {row.changelog?.trim() || "Changelog not available."}
      </Typography>
    </Box>
  );

  return (
    <Box>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          px: 2,
          pb: 1,
        }}
      >
        <Typography variant="h4" sx={{ lineHeight: 1.2 }}>
          Updates
        </Typography>
      </Box>

      <UpdateActions updates={updates} onComplete={refetch} />

      <UpdateList
        updates={updates}
        loading={isLoading}
        renderCollapseContent={renderCollapseContent}
      />
    </Box>
  );
};

export default UpdateStatus;

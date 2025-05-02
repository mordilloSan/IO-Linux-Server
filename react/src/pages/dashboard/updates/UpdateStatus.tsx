// components/system/UpdateStatus.tsx
import React, { useEffect, useState } from "react";
import { Box, Typography } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "@/utils/axios";
import UpdateActions from "./UpdateActions";
import UpdateList from "./UpdateList";
import { Update } from "@/types/update";
import ChangelogPanel from "./ChangelogPanel";

const UpdateStatus: React.FC = () => {
  const [isUpdating, setIsUpdating] = useState(false);
  const {
    data,
    refetch,
  } = useQuery<{ updates: Update[] }>({
    queryKey: ["updateInfo"],
    queryFn: () => axios.get("/system/updates").then((res) => res.data),
    refetchInterval: 50000,
    enabled: !isUpdating,
  });

  const updates = data?.updates || [];

  const queryClient = useQueryClient();

useEffect(() => {
  if (updates.length > 0) {
    updates.forEach((update) => {
      if (update.name) {
        queryClient.prefetchQuery({
          queryKey: ["changelog", update.name],
          queryFn: () =>
            axios
              .get("/system/updates/changelog", {
                params: { package: update.name },
              })
              .then((res) => res.data),
          staleTime: 5 * 60 * 1000,
        });
      }
    });
  }
}, [updates, queryClient]);


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

      <UpdateActions updates={updates} onComplete={refetch}   isUpdating={isUpdating}  setIsUpdating={setIsUpdating}/>

      <UpdateList
        updates={updates}
        onUpdateClick={(pkg) => <ChangelogPanel packageName={pkg} />}
      />
    </Box>
  );
};

export default UpdateStatus;

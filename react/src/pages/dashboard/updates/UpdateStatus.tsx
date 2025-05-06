// components/system/UpdateStatus.tsx
import { Box } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import React, { useEffect, useMemo, useState } from "react";

import UpdateActions from "./UpdateActions";
import UpdateList from "./UpdateList";

import { Update } from "@/types/update";
import axios from "@/utils/axios";

const UpdateStatus: React.FC = () => {
  const [isUpdating, setIsUpdating] = useState(false);
  const { data, isLoading, refetch } = useQuery<{ updates: Update[] }>({
    queryKey: ["updateInfo"],
    queryFn: () => axios.get("/system/updates").then((res) => res.data),
    refetchInterval: 50000,
    enabled: !isUpdating,
  });

  const updates = useMemo(() => data?.updates || [], [data]);

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
      <UpdateActions
        updates={updates}
        onComplete={refetch}
        isUpdating={isUpdating}
        setIsUpdating={setIsUpdating}
      />

      <UpdateList
        updates={updates}
        onUpdateClick={async (pkg) => {
          await axios.post("/system/update", { package: pkg });
          refetch();
        }}
        isLoading={isLoading}
      />
    </Box>
  );
};

export default UpdateStatus;

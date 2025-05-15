// components/system/UpdateStatus.tsx
import { Box } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import React, { useEffect, useMemo } from "react";

import UpdateActions from "./UpdateActions";
import UpdateList from "./UpdateList";

import { usePackageUpdater } from "@/hooks/usePackageUpdater";
import { Update } from "@/types/update";
import axios from "@/utils/axios";

const UpdateStatus: React.FC = () => {
  const queryClient = useQueryClient();

  const {
    data,
    isLoading,
    refetch: refetchUpdates,
  } = useQuery<{ updates: Update[] }>({
    queryKey: ["updateInfo"],
    queryFn: () => axios.get("/system/updates").then((res) => res.data),
    enabled: true,
    refetchInterval: 50000,
  });

  const updates = useMemo(() => data?.updates || [], [data]);

  const { updateOne, updateAll, updatingPackage, progress } =
    usePackageUpdater(refetchUpdates);

  // Prefetch changelogs
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
        onUpdateAll={() => updateAll(updates.map((u) => u.name))}
        isUpdating={!!updatingPackage}
        currentPackage={updatingPackage}
        progress={progress}
      />

      <UpdateList
        updates={updates}
        onUpdateClick={updateOne}
        isUpdating={!!updatingPackage}
        currentPackage={updatingPackage}
        onComplete={refetchUpdates}
        isLoading={isLoading}
      />
    </Box>
  );
};

export default UpdateStatus;

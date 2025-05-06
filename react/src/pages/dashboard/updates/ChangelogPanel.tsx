// components/system/ChangelogPanel.tsx
import { Typography, Box } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import React from "react";

import axios from "@/utils/axios";

interface Props {
  packageName: string;
}

const ChangelogPanel: React.FC<Props> = ({ packageName }) => {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["changelog", packageName],
    queryFn: () =>
      axios
        .get("/system/updates/changelog", { params: { package: packageName } })
        .then((res) => res.data),
    staleTime: 5 * 60 * 1000, // cache for 5 min
    enabled: !!packageName,
  });

  let content = "Changelog not available.";

  if (isLoading) content = "Loading changelog...";
  else if (isError) content = "Failed to load changelog.";
  else if (data?.changelog?.trim()) content = data.changelog.trim();

  return (
    <Box sx={{ whiteSpace: "pre-wrap", fontSize: 14 }}>
      <Typography variant="body2" color="text.secondary">
        {content}
      </Typography>
    </Box>
  );
};

export default ChangelogPanel;

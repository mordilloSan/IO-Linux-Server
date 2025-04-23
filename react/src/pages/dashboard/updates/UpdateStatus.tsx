import { useQuery } from "@tanstack/react-query";
import React, { useState } from "react";
import { Typography, Box, Button, LinearProgress, Card } from "@mui/material";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CollapsibleTable from "@/components/tables/CollapsibleTable";
import Loader from "@/components/Loader";
import axios from "@/utils/axios";
import { CollapsibleColumn } from "@/types/collapsible";

interface Update {
  name: string;
  version: string;
  severity: string;
  changelog?: string;
  packages?: string[];
}

interface UpdateInfo {
  updates: Update[];
}

const columns: CollapsibleColumn[] = [
  { field: "name", headerName: "Name" },
  { field: "version", headerName: "Version" },
  { field: "severity", headerName: "Severity" },
];

const UpdateStatus: React.FC = () => {
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateProgress, setUpdateProgress] = useState(0);
  const [currentPackage, setCurrentPackage] = useState("");

  const {
    data: updateInfo,
    isLoading: loadingSystemInfo,
    refetch,
  } = useQuery<UpdateInfo>({
    queryKey: ["updateInfo"],
    queryFn: async () => {
      const res = await axios.get("/system/updates");
      return res.data;
    },
    refetchInterval: 50000,
    enabled: !isUpdating,
  });

  const handleUpdateAll = async () => {
    if (!updateInfo || isUpdating) return;

    const allPackages = updateInfo.updates.flatMap((u) => u.packages || []);
    setIsUpdating(true);
    setUpdateProgress(0);
    setCurrentPackage("");

    const totalPackages = allPackages.length;

    for (let i = 0; i < totalPackages; i++) {
      const packageName = allPackages[i];
      setCurrentPackage(packageName);

      try {
        await axios.post("/api/update/", { packageName });
        setUpdateProgress(((i + 1) / totalPackages) * 100);
      } catch (error: any) {
        console.error(
          `Error updating package ${packageName}: ${error.message}`
        );
        continue;
      }
    }

    setIsUpdating(false);
    setCurrentPackage("");
    refetch();
  };

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
      {isUpdating && (
        <Box sx={{ textAlign: "center" }}>
          <Typography variant="h6" gutterBottom>
            Updating {currentPackage}...
          </Typography>
          <LinearProgress variant="determinate" value={updateProgress} />
          <Typography variant="body2" sx={{ mt: 1 }}>
            {`${Math.round(updateProgress)}% completed`}
          </Typography>
        </Box>
      )}

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
        {updates.length > 0 && (
          <Button
            variant="contained"
            color="primary"
            onClick={handleUpdateAll}
            disabled={isUpdating}
            sx={{ ml: 2, alignSelf: "center" }}
          >
            {isUpdating ? "Updating" : "Install All Updates"}
          </Button>
        )}
      </Box>

      {loadingSystemInfo ? (
        <Box sx={{ padding: 2 }}>
          <Card>
            <Box sx={{ py: 2.8 }}>
              <Loader />
            </Box>
          </Card>
        </Box>
      ) : updates.length > 0 ? (
        <CollapsibleTable
          rows={updates}
          columns={columns}
          renderCollapseContent={renderCollapseContent}
        />
      ) : (
        <Box sx={{ padding: 2 }}>
          <Card>
            <Box sx={{ display: "flex", alignItems: "center", py: 5 }}>
              <CheckCircleIcon
                color="success"
                sx={{ ml: 9, mr: 8, fontSize: 22 }}
              />
              <Typography variant="body1" fontSize={15}>
                System is up to date
              </Typography>
            </Box>
          </Card>
        </Box>
      )}
    </Box>
  );
};

export default UpdateStatus;

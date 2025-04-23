// components/system/UpdateActions.tsx
import React, { useState } from "react";
import { Box, Typography, Button, LinearProgress } from "@mui/material";
import axios from "@/utils/axios";
import { Update } from "@/types/update";

interface Props {
  updates: Update[];
  onComplete: () => void;
}

const UpdateActions: React.FC<Props> = ({ updates, onComplete }) => {
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateProgress, setUpdateProgress] = useState(0);
  const [currentPackage, setCurrentPackage] = useState("");

  const allPackages = updates.flatMap((u) => u.packages || []);
  const totalPackages = allPackages.length;

  const handleUpdateAll = async () => {
    if (totalPackages === 0 || isUpdating) return;

    setIsUpdating(true);
    setUpdateProgress(0);
    setCurrentPackage("");

    for (let i = 0; i < totalPackages; i++) {
      const pkg = allPackages[i];
      setCurrentPackage(pkg);

      try {
        await axios.post("/system/update", { package: pkg });
        setUpdateProgress(((i + 1) / totalPackages) * 100);
      } catch (err: any) {
        console.error(`Failed to update ${pkg}:`, err.message);
      }
    }

    setIsUpdating(false);
    setCurrentPackage("");
    onComplete();
  };

  return (
    <>
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
      <Box sx={{ display: "flex", justifyContent: "flex-end", pb: 2, px: 2 }}>
        <Button
          variant="contained"
          color="primary"
          onClick={handleUpdateAll}
          disabled={isUpdating}
        >
          {isUpdating ? "Updating" : "Install All Updates"}
        </Button>
      </Box>
    </>
  );
};

export default UpdateActions;

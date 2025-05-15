import {
  Box,
  CardContent,
  Typography,
  Chip,
  Grid,
  Collapse,
  CircularProgress,
} from "@mui/material";
import React, { useEffect, useRef, useState } from "react";

import ChangelogPanel from "./ChangelogPanel";

import FrostedCard from "@/components/cards/RootCard";
import { Update } from "@/types/update";

interface Props {
  updates: Update[];
  onUpdateClick: (pkg: string) => Promise<void>;
  isUpdating?: boolean;
  currentPackage?: string | null;
  onComplete: () => void;
  isLoading?: boolean;
}

const UpdateList: React.FC<Props> = ({
  updates,
  onUpdateClick,
  isUpdating,
  currentPackage,
  onComplete,
}) => {
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const toggleExpanded = (index: number) => {
    setExpandedIdx(index === expandedIdx ? null : index);
  };

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setExpandedIdx(null);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  if (!updates.length && !isUpdating) {
    return (
      <Box sx={{ textAlign: "left" }}>
        <Typography variant="h6">Your system is up to date 🎉</Typography>
      </Box>
    );
  }

  return (
    <Grid container spacing={2} sx={{ px: 2, pb: 2 }} ref={containerRef}>
      {updates.map((update, idx) => (
        <Grid key={idx} size={{ xs: 12, sm: 4, md: 4, lg: 3, xl: 2 }}>
          <FrostedCard
            variant="outlined"
            sx={{
              transition: "transform 0.2s, box-shadow 0.2s",
              "&:hover": {
                transform: "translateY(-4px)",
                boxShadow: "0 8px 24px rgba(0,0,0,0.35)",
              },
            }}
          >
            <CardContent>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  mb: 1,
                }}
              >
                <Typography
                  variant="h6"
                  title={update.name}
                  sx={{
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    maxWidth: "75%",
                  }}
                >
                  {update.name}
                </Typography>
                <Chip
                  label={update.severity}
                  size="small"
                  sx={{ backgroundColor: "transparent" }}
                />
              </Box>

              <Typography variant="body2" color="text.secondary" gutterBottom>
                Version: {update.version}
              </Typography>

              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1, mt: 1 }}>
                <Chip
                  label="View Changelog"
                  size="small"
                  variant="outlined"
                  onClick={() => toggleExpanded(idx)}
                  sx={{ cursor: "pointer" }}
                />
                <Chip
                  label={
                    currentPackage === update.name ? (
                      <CircularProgress size={16} />
                    ) : (
                      "Update"
                    )
                  }
                  size="small"
                  variant="outlined"
                  disabled={!!isUpdating}
                  onClick={async () => {
                    await onUpdateClick(update.name);
                    onComplete();
                  }}
                  sx={{ cursor: "pointer" }}
                />
              </Box>

              <Collapse in={expandedIdx === idx} timeout="auto" unmountOnExit>
                <Box sx={{ mt: 2 }}>
                  <ChangelogPanel packageName={update.name} />
                </Box>
              </Collapse>
            </CardContent>
          </FrostedCard>
        </Grid>
      ))}
    </Grid>
  );
};

export default UpdateList;

import React, { useState } from "react";
import {
  Box,
  CardContent,
  Typography,
  Chip,
  Stack,
  Collapse,
} from "@mui/material";
import { Update } from "@/types/update";
import FrostedCard from "@/components/cards/FrostedCard";
import ChangelogPanel from "./ChangelogPanel";

interface Props {
  updates: Update[];
  onUpdateClick: (pkg: string) => void;
}

const UpdateList: React.FC<Props> = ({ updates, onUpdateClick }) => {
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null);

  if (!updates.length) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <Typography variant="h6">Your system is up to date 🎉</Typography>
      </Box>
    );
  }

  const toggleExpanded = (index: number) => {
    setExpandedIdx(index === expandedIdx ? null : index);
  };

  return (
    <Stack spacing={2} sx={{ px: 2, pb: 2 }}>
      {updates.map((update, idx) => (
        <FrostedCard
          key={idx}
          variant="outlined"
          sx={{ width: "100%", maxWidth: 500 }}
        >
          <CardContent>
            <Box sx={{ display: "flex", justifyContent: "space-between", mb: 1 }}>
              <Typography variant="h6">{update.name}</Typography>
              <Chip
                label={update.severity}
                size="small"
                sx={{
                  backgroundColor: "transparent",
                }}
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
              {update.name && (
                <Chip
                  label="Update"
                  size="small"
                  variant="outlined"
                  onClick={() => onUpdateClick(update.name)}
                  sx={{ cursor: "pointer" }}
                />
              )}
            </Box>

            <Collapse in={expandedIdx === idx} timeout="auto" unmountOnExit>
              <Box sx={{ mt: 2 }}>
                <ChangelogPanel packageName={update.name} />
              </Box>
            </Collapse>
          </CardContent>
        </FrostedCard>
      ))}
    </Stack>
  );
};

export default UpdateList;

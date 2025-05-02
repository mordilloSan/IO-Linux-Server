import React, { useState, useContext } from "react";
import {
  Box,
  Fade,
  Paper,
  ToggleButton,
  ToggleButtonGroup,
} from "@mui/material";
import UpdateStatus from "./UpdateStatus";
import UpdateHistory from "./UpdateHistory";
import { ThemeContext } from "@/contexts/ThemeContext";

const Updates: React.FC = () => {
  const [tab, setTab] = useState<"updates" | "history" | "settings">("updates");
  const { primaryColor, isLoaded } = useContext(ThemeContext);

  // ❗ Block rendering until theme is loaded
  if (!isLoaded) return null;

  return (
    <Box sx={{ px: 2 }}>
      <Paper
        elevation={0}
        sx={{
          display: "flex",
          justifyContent: "center",
          p: 0.5,
          width: "fit-content",
          borderRadius: "999px",
          backgroundColor: "transparent",
          backdropFilter: "none",
        }}
      >
        <ToggleButtonGroup
          value={tab}
          exclusive
          onChange={(_, newTab) => newTab && setTab(newTab)}
          size="small"
          sx={{
            "& .MuiToggleButton-root": {
              color: "text.secondary",
              border: "none",
              borderRadius: "999px",
              px: 2,
            },
            "& .Mui-selected": {
              backgroundColor: `${primaryColor} !important`,
              color: "#fff",
              "&:hover": {
                backgroundColor: primaryColor,
              },
            },
          }}
        >
          <ToggleButton value="updates">Updates</ToggleButton>
          <ToggleButton value="history">History</ToggleButton>
          <ToggleButton value="settings">Settings</ToggleButton>
        </ToggleButtonGroup>
      </Paper>

      <Box sx={{ position: "relative", minHeight: 400 }}>
        <Fade in={tab === "updates"} timeout={300} unmountOnExit={false}>
          <Box
            sx={{
              display: tab === "updates" ? "block" : "none",
              position: "absolute",
              width: "100%",
            }}
          >
            <UpdateStatus />
          </Box>
        </Fade>

        <Fade in={tab === "history"} timeout={300} unmountOnExit={false}>
          <Box
            sx={{
              display: tab === "history" ? "block" : "none",
              position: "absolute",
              width: "100%",
            }}
          >
            <UpdateHistory />
          </Box>
        </Fade>

        <Fade in={tab === "settings"} timeout={300} unmountOnExit={false}>
          <Box
            sx={{
              display: tab === "settings" ? "block" : "none",
              position: "absolute",
              width: "100%",
            }}
          >
            <UpdateStatus />
          </Box>
        </Fade>
      </Box>
    </Box>
  );
};

export default Updates;

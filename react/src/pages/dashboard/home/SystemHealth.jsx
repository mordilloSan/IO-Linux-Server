"use client";
import React from "react";
import { useQuery } from "@tanstack/react-query";
import { Typography, Box, useTheme } from "@mui/material";
import CheckCircleOutlineIcon from "@mui/icons-material/CheckCircleOutline";
import HighlightOffIcon from "@mui/icons-material/HighlightOff";
import SecurityUpdateWarningIcon from "@mui/icons-material/SecurityUpdateWarning";
import CardWithBorder from "@/components/cards/CardWithBorder";
import RouterLink from "@/components/RouterLink";
import { useAuthenticatedFetch } from "@/utils/customFetch";

const SystemHealth = () => {
  const theme = useTheme();
  const customFetch = useAuthenticatedFetch();

  const { data: SystemHealth } = useQuery({
    queryKey: ["SystemHealth"],
    queryFn: () => customFetch("/api/updates/status"),
    refetchInterval: 50000,
  });

  const { data: systemStatus } = useQuery({
    queryKey: ["SystemStatus"],
    queryFn: () => customFetch("/api/system-status/status"),
    refetchInterval: 50000,
  });

  const { data: distroInfo } = useQuery({
    queryKey: ["DistroInfo"],
    queryFn: () => customFetch("/api/systeminfo/os"),
    refetchInterval: 50000,
  });

  const updates = SystemHealth?.updates || [];
  const units = systemStatus?.units || 0;
  const failed = systemStatus?.failed || 0;
  const distro = distroInfo?.os?.distro || "Unknown";

  // Determine status based on the number of failed units and available updates
  let statusColor = "green";
  let IconComponent = CheckCircleOutlineIcon;
  let iconLink = "/updates";

  if (failed > 0) {
    statusColor = "red";
    IconComponent = HighlightOffIcon;
    iconLink = "/services";
  } else if (updates.length > 0) {
    statusColor = theme.palette.warning.main;
    IconComponent = SecurityUpdateWarningIcon;
  }

  const stats = (
    <Box
      sx={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: 120,
        height: 120,
        borderRadius: "50%",
        color: statusColor,
      }}
    >
      <RouterLink href={iconLink}>
        <IconComponent sx={{ fontSize: 80 }} />
      </RouterLink>
    </Box>
  );

  const stats2 = (
    <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
      <Typography variant="body1">
        <Box component="span">
          <strong>Distro:</strong>
        </Box>{" "}
        {distro}
      </Typography>
      <Typography variant="body1">
        <RouterLink href="/updates">
          <strong>Updates:</strong>{" "}
          {updates.length > 0
            ? `${updates.length} available`
            : "None available"}
        </RouterLink>
      </Typography>
      <Typography variant="body1">
        <RouterLink href="/services">
          <strong>Services: </strong>
          {failed > 0 ? `${failed} failed` : `${units} running`}
        </RouterLink>
      </Typography>
    </Box>
  );

  return (
    <CardWithBorder
      title="System Health"
      stats={stats}
      stats2={stats2}
      avatarIcon={`simple-icons:${distroInfo?.os?.logofile || "linux"}`}
    />
  );
};

export default SystemHealth;

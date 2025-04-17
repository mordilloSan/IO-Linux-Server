import React from "react";
import { useQuery } from "@tanstack/react-query";
import { Typography, Box, useTheme, Link } from "@mui/material";
import CheckCircleOutlineIcon from "@mui/icons-material/CheckCircleOutline";
import HighlightOffIcon from "@mui/icons-material/HighlightOff";
import SecurityUpdateWarningIcon from "@mui/icons-material/SecurityUpdateWarning";
import CardWithBorder from "@/components/cards/CardWithBorder";
import axios from "@/utils/axios";

const SystemHealth = () => {
  const theme = useTheme();

  const { data: systemHealth } = useQuery({
    queryKey: ["SystemHealth"],
    queryFn: () => axios.get("/updates/status").then((res) => res.data),
    refetchInterval: 50000,
  });

  const { data: systemStatus } = useQuery({
    queryKey: ["SystemStatus"],
    queryFn: () => axios.get("/services/status").then((res) => res.data),
    refetchInterval: 50000,
  });

  const { data: distroInfo } = useQuery({
    queryKey: ["DistroInfo"],
    queryFn: () => axios.get("/system/info").then((res) => res.data),
    refetchInterval: 50000,
  });

  const updates = systemHealth?.updates || [];
  const units = systemStatus?.units || 0;
  const failed = systemStatus?.failed || 0;
  const distro = distroInfo?.platform || "Unknown";

  // Status logic
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
      <Link href={iconLink} underline="none">
        <IconComponent sx={{ fontSize: 80 }} />
      </Link>
    </Box>
  );

  const stats2 = (
    <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
      <Typography variant="body1">
        <strong>Distro:</strong> {distro}
      </Typography>
      <Typography variant="body1">
        <Link href="/updates" underline="hover">
          <strong>Updates:</strong>{" "}
          {updates.length > 0
            ? `${updates.length} available`
            : "None available"}
        </Link>
      </Typography>
      <Typography variant="body1">
        <Link href="/services" underline="hover">
          <strong>Services:</strong>{" "}
          {failed > 0 ? `${failed} failed` : `${units} running`}
        </Link>
      </Typography>
    </Box>
  );

  return (
    <CardWithBorder
      title="System Health"
      stats={stats}
      stats2={stats2}
      avatarIcon={`simple-icons:${distroInfo?.platform || "linux"}`}
    />
  );
};

export default SystemHealth;

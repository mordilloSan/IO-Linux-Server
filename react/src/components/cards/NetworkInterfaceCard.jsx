"use client";

import React from "react";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import RouterLink from "@/components/RouterLink";
import { styled } from "@mui/material/styles";

// Define a styled card component to include hover styles
const HoverableCard = styled(Card)(({ theme, color = "primary" }) => ({
  transition:
    "border 0.3s ease-in-out, box-shadow 0.3s ease-in-out, margin 0.3s ease-in-out",
  borderBottomWidth: "2px",
  borderBottomColor:
    theme.palette[color]?.darkerOpacity || theme.palette.primary.darkerOpacity,
  "&:hover": {
    borderBottomWidth: "3px",
    borderBottomColor: theme.palette[color]?.main || theme.palette.primary.main,
    boxShadow: theme.shadows[10],
    marginBlockEnd: "-1px",
  },
}));

const NetworkInterfaceCard = ({ name, ipAddress, tx, rx, carrierSpeed }) => {
  return (
    <HoverableCard sx={{ m: 1 }}>
      <CardContent>
        <Typography variant="h6" component="div">
          <RouterLink href={`/network/${name}`} passHref>
            {name}
          </RouterLink>
        </Typography>
        <Box sx={{ mt: 1 }}>
          <Typography variant="body2">
            <strong>IP Address:</strong> {ipAddress}
          </Typography>
          <Typography variant="body2">
            <strong>Tx:</strong> {tx}
          </Typography>
          <Typography variant="body2">
            <strong>Rx:</strong> {rx}
          </Typography>
          <Typography variant="body2">
            <strong>Carrier Speed:</strong> {carrierSpeed}
          </Typography>
        </Box>
      </CardContent>
    </HoverableCard>
  );
};

export default NetworkInterfaceCard;

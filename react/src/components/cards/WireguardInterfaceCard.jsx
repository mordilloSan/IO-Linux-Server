"use client";

import React from "react";
import { Card, CardContent, Typography, Box, IconButton } from "@mui/material";
import { Delete, Edit } from "@mui/icons-material";
import PowerSettingsNewIcon from "@mui/icons-material/PowerSettingsNew";
import { motion } from "framer-motion";
import { useTheme } from "@mui/material/styles";

const InterfaceCard = ({
  iface,
  selectedInterface,
  selectedCardRef,
  primaryColor,
  handleSelectInterface,
  handleToggleInterface,
  handleDelete,
  handleAddPeer,
}) => {
  const theme = useTheme();
  const color = "primary"; // Default color

  const hoverStyles = {
    borderBottomWidth: "3px",
    borderBottomColor: theme.palette[color]?.main || theme.palette.primary.main,
    boxShadow: theme.shadows[10],
    marginBlockEnd: "-1px",
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: 20 }}
      transition={{ duration: 0.3 }}
      layout
    >
      <Card
        ref={iface.name === selectedInterface ? selectedCardRef : null}
        sx={{
          cursor: "pointer",
          borderBottomWidth: "2px",
          borderBottomStyle: "solid",
          borderBottomColor:
            theme.palette[color]?.darkerOpacity ||
            theme.palette.primary.darkerOpacity,
          transition:
            "border 0.3s ease-in-out, box-shadow 0.3s ease-in-out, margin 0.3s ease-in-out",
          "&:hover": hoverStyles,
          ...(iface.name === selectedInterface && hoverStyles),
        }}
        onClick={() => handleSelectInterface(iface)}
      >
        <CardContent>
          <Box
            display="flex"
            justifyContent="space-between"
            alignItems="center"
          >
            <Typography variant="h6" sx={{ fontSize: "1.1rem" }}>
              {iface.name}
            </Typography>
            <Box>
              <IconButton
                sx={{
                  color: iface.isConnected === "Active" ? primaryColor : "gray",
                }}
                aria-label="Power"
                onClick={(e) => {
                  e.stopPropagation();
                  handleToggleInterface(
                    iface.name,
                    iface.isConnected === "Active" ? "down" : "up"
                  );
                }}
              >
                <PowerSettingsNewIcon />
              </IconButton>
              <IconButton
                onClick={(e) => {
                  e.stopPropagation();
                  handleDelete(iface.name);
                }}
              >
                <Delete />
              </IconButton>
              <IconButton
                onClick={(e) => {
                  e.stopPropagation();
                  handleAddPeer(iface.name, "peerName"); // Replace "peerName" with actual value
                }}
              >
                <Edit />
              </IconButton>
            </Box>
          </Box>
          <Typography variant="body2" color="textSecondary">
            Address: {iface.address}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Port: {iface.port}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Peers: {iface.peerCount}
          </Typography>
        </CardContent>
      </Card>
    </motion.div>
  );
};

export default InterfaceCard;

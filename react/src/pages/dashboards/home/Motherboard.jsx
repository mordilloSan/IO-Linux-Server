"use client";

import React from "react";

import { useQuery } from "@tanstack/react-query";
import { Typography, Box } from "@mui/material";
import CardWithBorder from "@/components/cards/CardWithBorder";
import TemperatureIcon from "@mui/icons-material/Thermostat";
import { useAuthenticatedFetch } from "@/utils/customFetch";

const MotherBoardInfo = () => {
  const customFetch = useAuthenticatedFetch();
  const { data: motherboardInfo } = useQuery({
    queryKey: ["motherboardInfo"],
    queryFn: () => customFetch("/api/systeminfo/baseboard"),
    refetchInterval: 50000,
  });

  const visibleDetails = motherboardInfo ? (
    <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
      <Typography variant="body1">{`${motherboardInfo?.baseboard?.manufacturer} - ${motherboardInfo?.baseboard?.model}`}</Typography>
      <Typography variant="body1">{`${motherboardInfo?.bios?.vendor}, V.${motherboardInfo?.bios?.version}`}</Typography>
    </Box>
  ) : (
    "No system information available."
  );

  const IconText = motherboardInfo
    ? `${motherboardInfo?.temperatures?.socket[0]}°C`
    : "--°C";

  return (
    <CardWithBorder
      title="Motherboard"
      stats={visibleDetails}
      icon_text={IconText}
      icon={TemperatureIcon}
      iconProps={{ sx: { color: "grey" } }}
      avatarIcon="bi:motherboard"
    />
  );
};

export default MotherBoardInfo;

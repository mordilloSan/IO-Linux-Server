import React from "react";
import { Grid } from "@mui/material";
import ErrorBoundary from "@/components/ErrorBoundary";

import Memory from "./Memory";
import SystemHealth from "./SystemHealth";
import FileSystem from "./FileSystem";
import Processor from "./Processor";
import NetworkInterfacesCard from "./NetworkInfo";
import MotherBoardInfo from "./MotherBoardInfo";
import GpuInfo from "./GpuInfo";

const cards = [
  { id: "system", component: <SystemHealth /> },
  { id: "cpu", component: <Processor /> },
  { id: "memory", component: <Memory /> },
  { id: "fs", component: <FileSystem /> },
  { id: "nic", component: <NetworkInterfacesCard /> },
  { id: "mb", component: <MotherBoardInfo /> },
  { id: "gpu", component: <GpuInfo /> },
];

const Dashboard: React.FC = () => {
  return (
    <Grid container spacing={4}>
      {cards.map((card) => (
        <ErrorBoundary key={card.id}>
          <Grid size={{ xs: 12, sm: 6, md: 6, lg: 4, xl: 3 }}>
            {card.component}
          </Grid>
        </ErrorBoundary>
      ))}
    </Grid>
  );
};

export default Dashboard;

import { Grid } from "@mui/material";
import React from "react";

import FileSystem from "./FileSystem";
import GpuInfo from "./GpuInfo";
import Memory from "./Memory";
import MotherBoardInfo from "./MotherBoardInfo";
import NetworkInterfacesCard from "./NetworkInfo";
import Processor from "./Processor";
import SystemHealth from "./SystemHealth";

import ErrorBoundary from "@/components/ErrorBoundary";

const MemoSystemHealth = React.memo(SystemHealth);
const MemoProcessor = React.memo(Processor);
const MemoMemory = React.memo(Memory);
const MemoFileSystem = React.memo(FileSystem);
const MemoNetworkInterfacesCard = React.memo(NetworkInterfacesCard);
const MemoMotherBoardInfo = React.memo(MotherBoardInfo);
const MemoGpuInfo = React.memo(GpuInfo);

const cards = [
  { id: "system", component: MemoSystemHealth },
  { id: "cpu", component: MemoProcessor },
  { id: "memory", component: MemoMemory },
  { id: "nic", component: MemoNetworkInterfacesCard },
  { id: "fs", component: MemoFileSystem },
  { id: "mb", component: MemoMotherBoardInfo },
  { id: "gpu", component: MemoGpuInfo },
];

const Dashboard: React.FC = () => {
  return (
    <Grid container spacing={4}>
      {cards.map(({ id, component: CardComponent }) => (
        <Grid key={id} size={{ xs: 12, sm: 6, md: 6, lg: 4, xl: 3 }}>
          <ErrorBoundary>
            <CardComponent />
          </ErrorBoundary>
        </Grid>
      ))}
    </Grid>
  );
};

export default Dashboard;

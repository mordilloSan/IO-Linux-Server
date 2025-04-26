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
  { id: "fs", component: MemoFileSystem },
  { id: "nic", component: MemoNetworkInterfacesCard },
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

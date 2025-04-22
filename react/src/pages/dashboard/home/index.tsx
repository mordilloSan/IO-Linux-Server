import React from "react";
import { Grid } from "@mui/material";
import Memory from "./Memory";
import SystemHealth from "./SystemHealth";
import FileSystem from "./FileSystem";
import Processor from "./Processor";

const Dashboard: React.FC = () => {
  return (
    <Grid container spacing={4}>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <SystemHealth />
      </Grid>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <Processor />
      </Grid>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <Memory />
      </Grid>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <FileSystem />
      </Grid>
    </Grid>
  );
};

export default Dashboard;

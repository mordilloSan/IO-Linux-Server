import React from "react";
import { Grid } from "@mui/material";
import Memory from "./Memory";
/* import SystemHealth from "./SystemHealth"; */
import FileSystem from "./FileSystem";
import Processor from "./Processor";
import ErrorBoundary from "@/components/ErrorBoundary";

const Dashboard: React.FC = () => {
  return (
    <Grid container spacing={4}>
      {/*      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <ErrorBoundary>
          <SystemHealth />
        </ErrorBoundary>
      </Grid>*/}
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <ErrorBoundary>
          <Processor />
        </ErrorBoundary>
      </Grid>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <ErrorBoundary>
          <Memory />
        </ErrorBoundary>
      </Grid>
      <Grid size={{ xs: 12, md: 6, xl: 4 }}>
        <ErrorBoundary>
          <FileSystem />
        </ErrorBoundary>
      </Grid>
    </Grid>
  );
};

export default Dashboard;

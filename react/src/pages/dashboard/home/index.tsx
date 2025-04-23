import React from "react";
import { Grid } from "@mui/material";
import Memory from "./Memory";
import SystemHealth from "./SystemHealth";
import FileSystem from "./FileSystem";
import Processor from "./Processor";

const cards = [
  { id: "system", component: <SystemHealth /> },
  { id: "cpu", component: <Processor /> },
  { id: "memory", component: <Memory /> },
  { id: "fs", component: <FileSystem /> },
];

const Dashboard: React.FC = () => {
  return (
    <Grid container spacing={4}>
      {cards.map((card) => (
        <Grid key={card.id} size={{ xs: 12, sm: 6, md: 6, lg: 4, xl: 3 }}>
          {card.component}
        </Grid>
      ))}
    </Grid>
  );
};

export default Dashboard;

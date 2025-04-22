import React from "react";
import { Box } from "@mui/material";
import UpdateStatus from "./UpdateStatus";
import UpdateHistory from "./UpdateHistory";

const Updates: React.FC = () => {
  return (
    <Box>
      <UpdateStatus />
      <UpdateHistory/>
    </Box>
  );
};

export default Updates;

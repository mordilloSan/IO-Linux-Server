import React from "react";
import { Box, CircularProgress } from "@mui/material";

function ComponentLoader() {
  return (
    <Box
      sx={{
        justifyContent: "center",
        alignItems: "center",
        display: "flex",
        minHeight: "100%",
      }}
    >
      <CircularProgress color="secondary" />
    </Box>
  );
}

export default ComponentLoader;

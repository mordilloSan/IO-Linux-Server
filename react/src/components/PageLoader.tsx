import React from "react";
import { Box, useTheme, alpha } from "@mui/material";
import { motion } from "framer-motion";

const animation = {
  animate: {
    x: ["-150px", "300px"], // Start off-screen and move across
    transition: {
      duration: 1.0,
      repeat: Infinity,
      ease: "easeInOut",
    },
  },
};

function Loader() {
  const theme = useTheme();
  const color = theme.palette.primary.main;

  return (
    <Box
      sx={{
        width: "100%",
        height: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        background: "transparent",
      }}
    >
      <Box
        sx={{
          width: 300,
          height: 6,
          backgroundColor: "#e0e0e0",
          borderRadius: 3,
          overflow: "hidden",
          position: "relative",
        }}
      >
        <Box
          component={motion.div}
          variants={animation}
          animate="animate"
          sx={{
            height: "100%",
            width: 150, // wider than container for trail
            position: "absolute",
            left: 0,
            top: 0,
            background: `linear-gradient(90deg, ${color}, ${alpha(
              color,
              0.5
            )})`,
            filter: "blur(1px)",
            borderRadius: 3,
          }}
        />
      </Box>
    </Box>
  );
}

export default Loader;

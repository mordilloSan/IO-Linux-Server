import { Box, useTheme, alpha } from "@mui/material";
import { motion } from "framer-motion";
import React from "react";

const animation = {
  animate: {
    x: ["-150px", "300px"],
    transition: {
      duration: 1.0,
      repeat: Infinity,
      ease: "easeInOut",
    },
  },
};

function PageLoader() {
  const theme = useTheme();
  const color = theme.palette.primary.main;

  return (
    <Box
      sx={{
        position: "fixed",
        top: 0,
        left: 0,
        width: "100vw",
        height: "100vh",
        backgroundColor: theme.palette.background.default,
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        zIndex: 1300,
      }}
    >
      <Box
        sx={{
          width: 300,
          height: 6,
          backgroundColor: theme.palette.background.paper,
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
            width: 150,
            position: "absolute",
            left: 0,
            top: 0,
            background: `linear-gradient(90deg, ${color}, ${alpha(
              color,
              0.5,
            )})`,
            filter: "blur(1px)",
            borderRadius: 3,
          }}
        />
      </Box>
    </Box>
  );
}

export default PageLoader;

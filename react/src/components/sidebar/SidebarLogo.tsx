import React from "react";
import { Box, Typography, useTheme } from "@mui/material";
import { motion } from "framer-motion";
import useSidebar from "@/hooks/useSidebar";
import { collapsedDrawerWidth } from "@/constants";

const SidebarLogo: React.FC = () => {
  const theme = useTheme();
  const { collapsed, hovered, isDesktop } = useSidebar();
  const showText = !collapsed || (hovered && isDesktop);

  return (
    <Box display="flex" alignItems="center" minHeight={56}>
      <Typography
        variant="h6"
        noWrap
        sx={{
          fontWeight: 700,
          fontSize: "1.25rem",
          letterSpacing: "0.1rem",
          display: "inline-flex",
          alignItems: "center",
        }}
      >
        <motion.span
          initial={false}
          animate={{
            opacity: showText ? 1 : 0,
            marginRight: showText ? 8 : -55,
          }}
          transition={{
            duration: theme.transitions.duration.standard / 1000, // ms → seconds
            ease: "easeInOut",
          }}
          style={{ color: theme.palette.text.primary, display: "inline-block" }}
        >
          Linux
        </motion.span>
        <Box component="span" sx={{ color: theme.palette.primary.main }}>
          i/O
        </Box>
      </Typography>
    </Box>
  );
};

export default SidebarLogo;

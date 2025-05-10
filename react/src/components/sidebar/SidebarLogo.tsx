import React from "react";
import { Box } from "@mui/material";
import useSidebar from "@/hooks/useSidebar";
import LogoDisplay from "../logo/LogoDisplay";

const SidebarLogo: React.FC = () => {
  const { collapsed, hovered, isDesktop } = useSidebar();
  const showText = !collapsed || (hovered && isDesktop);

  return (
    <Box
      display="flex"
      alignItems="center"
      minHeight={56}
      justifyContent={
        collapsed && !hovered && isDesktop ? "center" : "flex-start"
      }
      pl={collapsed && !hovered && isDesktop ? 0 : 3}
    >
      <LogoDisplay showText={showText} />
    </Box>
  );
};

export default SidebarLogo;

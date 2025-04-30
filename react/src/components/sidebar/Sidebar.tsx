import React, { useState } from "react";
import { Drawer, Box, useTheme } from "@mui/material";
import { ReactComponent as Logo } from "@/assets/logo.svg";
import { SidebarItemsType } from "@/types/sidebar";
import SidebarNav from "./SidebarNav";
import { ChevronLeft, ChevronRight } from "@mui/icons-material";
import { collapsedDrawerWidth, drawerWidth } from "@/constants";

export type SidebarProps = {
  PaperProps: {
    style: {
      width: number;
    };
  };
  variant?: "permanent" | "persistent" | "temporary";
  open?: boolean;
  onClose?: () => void;
  items: SidebarItemsType[];
  collapsed?: boolean;
  onSidebarCollapseToggle?: () => void;
};

const Sidebar: React.FC<SidebarProps> = ({
  items,
  collapsed = false,
  onSidebarCollapseToggle,
  ...rest
}) => {
  const theme = useTheme();
  const [isHovered, setIsHovered] = useState(false); // Track hover state for hover expansion
  const [isCollapsed, setIsCollapsed] = useState(collapsed); // Track actual collapsed state

  // Set collapsed to false when hovered
  const handleMouseEnter = () => {
    setIsHovered(true);
    setIsCollapsed(false); // Expand on hover only if collapsed
  };

  // Reset collapsed back to true when mouse leaves
  const handleMouseLeave = () => {
    setIsHovered(false);
    setIsCollapsed(collapsed); // Reset to collapsed state
  };

  return (
    <Drawer
      key={theme.palette.mode}
      {...rest}
      slotProps={{
        paper: {
          sx: {
            width: isCollapsed && !isHovered ? collapsedDrawerWidth : drawerWidth, // Adjust width when collapsed and hovered
            borderRight: 0,
            backgroundColor: theme.sidebar.background,
            scrollbarWidth: "none",
            transition: theme.transitions.create(["width", "background-color"], {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.standard,
            }),
            overflowX: "hidden",
            "& > div": {
              borderRight: 0,
            },
          },
        },
      }}
      onMouseEnter={handleMouseEnter}  // Set hover state to true when hovering the sidebar
      onMouseLeave={handleMouseLeave} // Set hover state to false when mouse leaves
    >
      {/* --- Sidebar Header (just logo) --- */}
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: theme.sidebar.header.background,
          minHeight: { xs: 56, sm: 64 },
          px: 6,
          flexGrow: 0,
          position: "relative",
        }}
      >
        <Box
          component={Logo}
          sx={{
            fill: theme.palette.primary.main,
            height: 42,
            marginRight: 2,
          }}
        />
        {/* Collapse Sidebar button (only desktop) */}
        {onSidebarCollapseToggle && (
            <div
              color="inherit"
              aria-label="Collapse sidebar"
              onClick={onSidebarCollapseToggle}

            >
              {(collapsed && isHovered)? (
                <ChevronRight sx={{ width: 22, height: 22 }} />
              ) : (
                <ChevronLeft sx={{ width: 22, height: 22 }} />
              )}
            </div>
          )}
      </Box>

      {/* --- Sidebar Navigation --- */}
      <SidebarNav items={items} collapsed={isCollapsed} />
    </Drawer>
  );
};

export default Sidebar;

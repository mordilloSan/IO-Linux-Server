import React from "react";
import { Drawer, Box, useTheme } from "@mui/material";
import { ReactComponent as Logo } from "@/assets/logo.svg";
import { SidebarItemsType } from "@/types/sidebar";
import SidebarNav from "./SidebarNav";
import { ChevronLeft, ChevronRight } from "@mui/icons-material";
import { collapsedDrawerWidth, drawerWidth } from "@/constants";
import useSidebar from "@/hooks/useSidebar";

export type SidebarProps = {
  items: SidebarItemsType[];
};

const Sidebar: React.FC<SidebarProps> = ({ items }) => {
  const theme = useTheme();
  const {
    collapsed,
    hovered,
    setHovered,
    toggleCollapse,
    isDesktop,
    hoverEnabledRef,
    mobileOpen,
    setMobileOpen,
  } = useSidebar();

  const effectiveWidth = !isDesktop
    ? drawerWidth
    : collapsed && !hovered
    ? collapsedDrawerWidth
    : drawerWidth;

  const handleMouseEnter = () => {
    if (hoverEnabledRef.current) setHovered(true);
  };

  const handleMouseLeave = () => setHovered(false);

  return (
    <Drawer
      variant={isDesktop ? "permanent" : "temporary"}
      open={isDesktop ? true : mobileOpen}
      onClose={() => setMobileOpen(false)}
      slotProps={{
        paper: {
          sx: {
            width: effectiveWidth,
            borderRight: 0,
            backgroundColor: theme.sidebar.background,
            transition: theme.transitions.create(
              ["width", "background-color"],
              {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.standard,
              }
            ),
            overflowX: "hidden",
            "& > div": {
              borderRight: 0,
            },
          },
        },
      }}
      onMouseEnter={isDesktop ? handleMouseEnter : undefined}
      onMouseLeave={isDesktop ? handleMouseLeave : undefined}
    >
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: theme.sidebar.header.background,
          minHeight: { xs: 56, sm: 64 },
          px: 6,
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

        {isDesktop && (!collapsed || (hovered && collapsed)) && (
          <div
            onClick={toggleCollapse}
            style={{
              position: "absolute",
              right: 0,
              top: "50%",
              transform: "translateY(-50%)",
              cursor: "pointer",
              display: "inline-flex",
            }}
          >
            {!collapsed && <ChevronLeft sx={{ width: 22, height: 22 }} />}
            {hovered && collapsed && (
              <ChevronRight sx={{ width: 22, height: 22 }} />
            )}
          </div>
        )}
      </Box>

      <SidebarNav
        items={items}
        collapsed={isDesktop && collapsed && !hovered}
      />
    </Drawer>
  );
};

export default Sidebar;

import React from "react";
import { Drawer, Box, useTheme } from "@mui/material";
import { ReactComponent as Logo } from "@/assets/logo.svg";
import { SidebarItemsType } from "@/types/sidebar";
import SidebarNav from "./SidebarNav";

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
};

const Sidebar: React.FC<SidebarProps> = ({
  items,
  collapsed = false,
  ...rest
}) => {
  const theme = useTheme();

  return (
    <Drawer
      key={theme.palette.mode}
      {...rest}
      slotProps={{
        paper: {
          sx: {
            width: rest.PaperProps.style.width,
            borderRight: 0,
            backgroundColor: theme.sidebar.background,
            scrollbarWidth: "none",
            transition: theme.transitions.create(
              ["width", "background-color"],
              {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.standard,
              }
            ), // 👈 this is the key!
            overflowX: "hidden", // 👈 avoid horizontal scrollbar when collapsing
            "& > div": {
              borderRight: 0,
            },
          },
        },
      }}
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
      </Box>

      {/* --- Sidebar Navigation --- */}
      <SidebarNav items={items} collapsed={collapsed} />
    </Drawer>
  );
};

export default Sidebar;

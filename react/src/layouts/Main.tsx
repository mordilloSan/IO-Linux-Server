import React, { useEffect, useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import {
  Box,
  CssBaseline,
  Container,
  Paper,
  useMediaQuery,
} from "@mui/material";
import { useTheme } from "@mui/material/styles";

import Navbar from "@/components/navbar/Navbar";
import Sidebar from "@/components/sidebar/Sidebar";
import Footer from "@/components/Footer";
import ErrorBoundary from "@/components/ErrorBoundary";
import dashboardItems from "@/components/sidebar/dashboardItems";
import { drawerWidth, collapsedDrawerWidth } from "@/constants";

const Dashboard: React.FC = () => {
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up("md"));

  const handleDrawerToggle = () => {
    setMobileOpen((prev) => !prev);
  };

  const handleSidebarCollapseToggle = () => {
    setCollapsed((prev) => !prev);
  };

  useEffect(() => {
    setMobileOpen(false);
  }, [location.pathname]);

  const sidebarWidth = isDesktop
    ? collapsed
      ? collapsedDrawerWidth
      : drawerWidth
    : drawerWidth;

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <CssBaseline />

      {/* Sidebar */}
      <Sidebar
        PaperProps={{ style: { width: sidebarWidth } }}
        variant={isDesktop ? "permanent" : "temporary"}
        open={mobileOpen}
        onClose={handleDrawerToggle}
        items={dashboardItems}
        collapsed={collapsed}
      />

      {/* Main Content */}
      <Box
  sx={{
    flex: 1,
    display: "flex",
    flexDirection: "column",
    width: "100%",
    minHeight: "100vh", // Important
    transition: theme.transitions.create(["margin-left", "width"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    ml: { md: `${sidebarWidth}px` },
  }}
>

        <Navbar
          onDrawerToggle={handleDrawerToggle}
          onSidebarCollapseToggle={handleSidebarCollapseToggle}
          collapsed={collapsed}
        />

<Paper
  elevation={0}
  sx={{
    flexGrow: 1, // 👈 not just flex: 1
    background: theme.palette.background.default,
    p: { xs: 5, lg: 7 },
    boxShadow: "none",
    minHeight: 0, // 👈 allows it to shrink correctly if needed
  }}
>
  <Container maxWidth={false} sx={{ height: "100%" }}>
            <ErrorBoundary>
              <Outlet />
            </ErrorBoundary>
          </Container>
        </Paper>

        <Footer />
      </Box>
    </Box>
  );
};

export default Dashboard;

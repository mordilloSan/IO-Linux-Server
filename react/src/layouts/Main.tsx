import React, { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { Box, CssBaseline } from "@mui/material";
import { useTheme } from "@mui/material/styles";

import Navbar from "@/components/navbar/Navbar";
import Sidebar from "@/components/sidebar/Sidebar";
import Footer from "@/components/Footer";
import ErrorBoundary from "@/components/ErrorBoundary";
import dashboardItems from "@/components/sidebar/dashboardItems";
import useSidebar from "@/hooks/useSidebar";
import useAppTheme from "@/hooks/useAppTheme";

const Dashboard: React.FC = () => {
  const location = useLocation();
  const theme = useTheme();
  const { isLoaded } = useAppTheme();
  const {
    mobileOpen,
    toggleMobileOpen,
    setMobileOpen,
    sidebarWidth,
    isDesktop,
  } = useSidebar();


  // Wait for theme to load before rendering layout
  if (!isLoaded) return null;

  // Auto-close mobile drawer on route change
  useEffect(() => {
    setMobileOpen(false);
  }, [location.pathname, setMobileOpen]);

  return (
    <Box sx={{ display: "flex", height: "100vh" }}>
      <CssBaseline />

      {/* Sidebar */}
      <Sidebar
        items={dashboardItems}
      />

      {/* Main content */}
      <Box
        sx={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          transition: theme.transitions.create(["margin-left", "width"], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
          ml: { md: `${sidebarWidth}px` },
        }}
      >
        <Navbar onDrawerToggle={toggleMobileOpen} />

        {/* Scrollable Content Area */}
        <Box
          className="custom-scrollbar"
          sx={{
            flex: 1,
            overflow: "auto",
            background: theme.palette.background.default,
            p: { xs: 5, lg: 7 },
          }}
        >
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </Box>

        {/* Footer Always at Bottom */}
        <Footer />
      </Box>
    </Box>
  );
};

export default Dashboard;

import React, {
  createContext,
  useState,
  useCallback,
  useMemo,
  useEffect,
  useRef,
} from "react";
import { useMediaQuery, useTheme } from "@mui/material";
import { drawerWidth, collapsedDrawerWidth } from "@/constants";

export interface SidebarContextType {
  collapsed: boolean;
  hovered: boolean;
  mobileOpen: boolean;
  isDesktop: boolean;
  sidebarWidth: number;
  setHovered: (value: boolean) => void;
  setMobileOpen: (value: boolean) => void;
  toggleCollapse: () => void;
  toggleMobileOpen: () => void;
  hoverEnabledRef: React.RefObject<boolean>;
}

export const SidebarContext = createContext<SidebarContextType | undefined>(
  undefined
);

export const SidebarProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up("md"));

  const [collapsed, setCollapsed] = useState(true);
  const [hovered, setHovered] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const hoverEnabled = useRef(true);
  const desktopCollapsed = useRef(true); // store previous desktop collapse state

  const toggleCollapse = useCallback(() => {
    setCollapsed((prev) => {
      const newState = !prev;
      desktopCollapsed.current = newState;
  
      if (isDesktop && newState === true) {
        // User just collapsed the sidebar — suppress hover briefly
        hoverEnabled.current = false;
        setHovered(false);
        setTimeout(() => {
          hoverEnabled.current = true;
        }, 200);
      }
  
      return newState;
    });
  }, [isDesktop]);
  

  const toggleMobileOpen = useCallback(() => {
    if (!isDesktop) return; // collapse shouldn't affect mobile
    setMobileOpen((prev) => !prev);
  }, []);

  // Handle screen size changes
  useEffect(() => {
    if (isDesktop) {
      // Restore collapsed state and reset mobile drawer
      setCollapsed(desktopCollapsed.current);
      setMobileOpen(false);
    } else {
      // Store current desktop state, and fully expand sidebar on mobile
      desktopCollapsed.current = collapsed;
      setCollapsed(false);
      setHovered(false);
      setMobileOpen(false); // or true, if you want it open on mobile by default
    }
  }, [isDesktop]);
  

  const sidebarWidth = useMemo(() => {
    return isDesktop
      ? collapsed
        ? collapsedDrawerWidth
        : drawerWidth
      : drawerWidth;
  }, [isDesktop, collapsed]);

  const value = useMemo(
    () => ({
      collapsed,
      hovered,
      mobileOpen,
      isDesktop,
      sidebarWidth,
      setHovered,
      setMobileOpen,
      toggleCollapse,
      toggleMobileOpen,
      hoverEnabledRef: hoverEnabled,
    }),
    [
      collapsed,
      hovered,
      mobileOpen,
      isDesktop,
      sidebarWidth,
      setHovered,
      setMobileOpen,
      toggleCollapse,
      toggleMobileOpen,
    ]
  );

  return (
    <SidebarContext.Provider value={value}>{children}</SidebarContext.Provider>
  );
};

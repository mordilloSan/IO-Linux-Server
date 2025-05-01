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
}

export const SidebarContext = createContext<SidebarContextType | undefined>(
  undefined
);

export const SidebarProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up("md"));

  const [collapsed, setCollapsed] = useState(true); // default collapsed on desktop
  const [hovered, setHovered] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const desktopCollapsed = useRef(true); // store previous desktop collapse state

  const toggleCollapse = useCallback(() => {
    setCollapsed((prev) => {
      desktopCollapsed.current = !prev; // update memory
      return !prev;
    });
  }, []);

  const toggleMobileOpen = useCallback(() => {
    setMobileOpen((prev) => !prev);
  }, []);

  // Handle screen size changes
  useEffect(() => {
    if (isDesktop) {
      setCollapsed(desktopCollapsed.current); // restore previous desktop state
    } else {
      // save desktop state and force expand for mobile
      desktopCollapsed.current = collapsed;
      setCollapsed(false);
      setHovered(false);
      setMobileOpen(false); // always close drawer on mobile
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

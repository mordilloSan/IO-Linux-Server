import React, {
  createContext,
  useState,
  useCallback,
  useMemo,
  useEffect,
  useRef,
} from "react";
import { useMediaQuery, useTheme as useMuiTheme } from "@mui/material";
import { drawerWidth, collapsedDrawerWidth } from "@/constants";
import useAppTheme from "@/hooks/useAppTheme";

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
  const muiTheme = useMuiTheme();
  const isDesktop = useMediaQuery(muiTheme.breakpoints.up("md"));
  const { sidebarColapsed: collapsed, setSidebarColapsed } = useAppTheme();

  const [hovered, setHovered] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const hoverEnabled = useRef(true);
  const desktopCollapsed = useRef<boolean>(collapsed);

  const toggleCollapse = useCallback(() => {
    setSidebarColapsed((prev) => {
      const newState = !prev;
      desktopCollapsed.current = newState;

      if (isDesktop && newState) {
        hoverEnabled.current = false;
        setHovered(false);
        setTimeout(() => {
          hoverEnabled.current = true;
        }, 200);
      }

      return newState;
    });
  }, [isDesktop, setSidebarColapsed]);

  const toggleMobileOpen = useCallback(() => {
    if (isDesktop) return;
    setMobileOpen((prev) => !prev);
  }, [isDesktop]);

  useEffect(() => {
    if (isDesktop) {
      desktopCollapsed.current = collapsed; // sync current state
      setMobileOpen(false);
    } else {
      setHovered(false);
      setMobileOpen(false);
    }
  }, [isDesktop, collapsed]);

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

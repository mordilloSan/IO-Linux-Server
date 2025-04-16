import React from "react";
import { IconButton, Tooltip } from "@mui/material";
import { Sun, Moon } from "lucide-react";

import { THEMES } from "@/constants";
import useTheme from "@/hooks/useTheme";

function NavbarThemeToggle() {
  const { theme, setTheme } = useTheme();
  const isDark = theme === THEMES.DARK;

  const toggleTheme = () => {
    setTheme(isDark ? THEMES.LIGHT : THEMES.DARK);
  };

  return (
    <Tooltip title={isDark ? "Switch to light mode" : "Switch to dark mode"}>
      <IconButton color="inherit" onClick={toggleTheme} size="large">
        {isDark ? <Sun /> : <Moon />}
      </IconButton>
    </Tooltip>
  );
}

export default NavbarThemeToggle;

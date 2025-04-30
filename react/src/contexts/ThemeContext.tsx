import React, { createContext, useEffect, useState, useCallback } from "react";
import { THEMES } from "@/constants";
import { ThemeContextType, ThemeProviderProps } from "@/types/theme";
import axios from "@/utils/axios";

const initialState: ThemeContextType = {
  theme: THEMES.DARK,
  setTheme: () => {},
  primaryColor: undefined,
  setPrimaryColor: () => {},
  toggleTheme: () => {},
};

const ThemeContext = createContext<ThemeContextType>(initialState);

function ThemeProvider({ children }: ThemeProviderProps) {
  const [theme, _setTheme] = useState<string>(initialState.theme);
  const [primaryColor, _setPrimaryColor] = useState<string | undefined>(
    undefined
  );

  // Fetch the theme from the backend on component mount
  useEffect(() => {
    const fetchTheme = async () => {
      try {
        const response = await axios.get("/theme/get");
        const fetchedTheme =
          response.data.theme === "LIGHT" ? THEMES.LIGHT : THEMES.DARK;
        _setTheme(fetchedTheme);
      } catch (error) {
        console.error("Error fetching theme from backend:", error);
        // Optionally, set a default theme if API fails
        _setTheme(THEMES.DARK);
      }
    };
    fetchTheme();
  }, []);

  const setTheme = useCallback((newTheme: string) => {
    _setTheme(newTheme);
    axios.post("/theme/set", { theme: newTheme }).catch((error) => {
      console.error("Error saving theme:", error);
    });
  }, []); // `setTheme` does not depend on any state, so it only needs to be created once

  const setPrimaryColor = useCallback((color: string) => {
    _setPrimaryColor(color);
  }, []); // `setPrimaryColor` does not depend on any state, so it only needs to be created once

  const toggleTheme = useCallback(() => {
    const newTheme = theme === THEMES.DARK ? THEMES.LIGHT : THEMES.DARK;
    setTheme(newTheme);
  }, [theme, setTheme]);

  return (
    <ThemeContext.Provider
      value={{ theme, setTheme, primaryColor, setPrimaryColor, toggleTheme }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export { ThemeProvider, ThemeContext };

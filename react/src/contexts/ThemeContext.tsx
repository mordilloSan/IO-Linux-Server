import React, { useEffect } from "react";
import { THEMES } from "@/constants";
import { ThemeContextType } from "@/types/theme";

const initialState: ThemeContextType = {
  theme: THEMES.DARK,
  setTheme: () => {},
  primaryColor: undefined,
  setPrimaryColor: () => {},
};

const ThemeContext = React.createContext<ThemeContextType>(initialState);

type ThemeProviderProps = {
  children: React.ReactNode;
};

function ThemeProvider({ children }: ThemeProviderProps) {
  const [theme, _setTheme] = React.useState<string>(initialState.theme);
  const [primaryColor, _setPrimaryColor] = React.useState<string | undefined>(
    undefined
  );

  useEffect(() => {
    const storedTheme = localStorage.getItem("theme");
    const storedColor = localStorage.getItem("primaryColor");

    if (storedTheme) _setTheme(JSON.parse(storedTheme));
    if (storedColor) _setPrimaryColor(JSON.parse(storedColor));
  }, []);

  const setTheme = (newTheme: string) => {
    localStorage.setItem("theme", JSON.stringify(newTheme));
    _setTheme(newTheme);
  };

  const setPrimaryColor = (color: string) => {
    localStorage.setItem("primaryColor", JSON.stringify(color));
    _setPrimaryColor(color);
  };

  return (
    <ThemeContext.Provider
      value={{ theme, setTheme, primaryColor, setPrimaryColor }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export { ThemeProvider, ThemeContext };

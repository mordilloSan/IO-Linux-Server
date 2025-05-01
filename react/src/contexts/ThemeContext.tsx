import React, {
  createContext,
  useEffect,
  useState,
  useCallback,
  useMemo,
} from "react";
import { DEFAULT_PRIMARY_COLOR, THEMES } from "@/constants";
import { ThemeContextType, ThemeProviderProps } from "@/types/theme";
import axios from "@/utils/axios";

const initialState: ThemeContextType = {
  theme: THEMES.DARK,
  setTheme: () => {},
  primaryColor: DEFAULT_PRIMARY_COLOR,
  setPrimaryColor: () => {},
  toggleTheme: () => {},
};

const ThemeContext = createContext<ThemeContextType>(initialState);

const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [theme, _setTheme] = useState<string>(initialState.theme);
  const [primaryColor, _setPrimaryColor] = useState<string>(DEFAULT_PRIMARY_COLOR);

  // Fetch theme and color on mount
  useEffect(() => {
    const fetchTheme = async () => {
      try {
        const response = await axios.get("/theme/get");
        const fetchedTheme = response.data.theme === "LIGHT" ? THEMES.LIGHT : THEMES.DARK;
        const fetchedColor = response.data.primaryColor;

        _setTheme(fetchedTheme);
        _setPrimaryColor(fetchedColor || DEFAULT_PRIMARY_COLOR);
      } catch (error) {
        console.error("Error fetching theme from backend:", error);
        _setTheme(THEMES.DARK);
        _setPrimaryColor(DEFAULT_PRIMARY_COLOR);
      }
    };

    fetchTheme();
  }, []);

  const saveThemeSettings = useCallback(
    (themeToSave: string, colorToSave: string) => {
      axios
        .post("/theme/set", {
          theme: themeToSave,
          primaryColor: colorToSave,
        })
        .catch((error) => {
          console.error("Error saving theme settings:", error);
        });
    },
    []
  );

  const setTheme = useCallback(
    (newTheme: string) => {
      _setTheme(newTheme);
      saveThemeSettings(newTheme, primaryColor);
    },
    [primaryColor, saveThemeSettings]
  );

  const setPrimaryColor = useCallback(
    (color: string) => {
      _setPrimaryColor(color);
      saveThemeSettings(theme, color);
    },
    [theme, saveThemeSettings]
  );

  const toggleTheme = useCallback(() => {
    const newTheme = theme === THEMES.DARK ? THEMES.LIGHT : THEMES.DARK;
    setTheme(newTheme);
  }, [theme, setTheme]);

  const contextValue = useMemo(
    () => ({
      theme,
      setTheme,
      primaryColor,
      setPrimaryColor,
      toggleTheme,
    }),
    [theme, setTheme, primaryColor, setPrimaryColor, toggleTheme]
  );

  return (
    <ThemeContext.Provider value={contextValue}>
      {children}
    </ThemeContext.Provider>
  );
};

export { ThemeProvider, ThemeContext };

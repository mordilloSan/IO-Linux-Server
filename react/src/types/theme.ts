import { Theme } from "@mui/material";

export type ThemeProps = {
  theme: Theme & { palette: any };
};

export type ThemeContextType = {
  theme: string;
  setTheme: (theme: string) => void;
  primaryColor: string | undefined;
  setPrimaryColor: (color: string) => void;
  toggleTheme: () => void; // Add toggleTheme to the context type
};

export type ThemeProviderProps = {
  children: React.ReactNode;
};

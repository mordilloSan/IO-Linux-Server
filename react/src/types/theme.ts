import { Theme } from "@mui/material";

export type ThemeProps = {
  theme: Theme & { palette: any };
};

export type ThemeContextType = {
  theme: string;
  setTheme: (theme: string) => void;
  primaryColor: string | undefined;
  setPrimaryColor: (color: string) => void;
};

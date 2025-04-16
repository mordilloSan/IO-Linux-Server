import { PaletteMode, Theme as MuiTheme, ThemeOptions } from "@mui/material/styles";

type MainContrastText = {
  main: string;
  contrastText: string;
};

export type VariantType = {
  name: string;
  palette: {
    mode: PaletteMode;
    primary: MainContrastText;
    secondary: MainContrastText;
    background: {
      default: string;
      paper: string;
    };
    text?: {
      primary?: string;
      secondary?: string;
    };
  };
  header: {
    color: string;
    background: string;
    search: { color: string };
    indicator: { background: string };
  };
  footer: {
    color: string;
    background: string;
  };
  sidebar: {
    color: string;
    background: string;
    header: {
      color: string;
      background: string;
      brand: { color: string };
    };
    footer: {
      color: string;
      background: string;
    };
    badge: {
      color: string;
      background: string;
    };
  };
};

export type ThemeProps = {
  theme: MuiTheme;
};

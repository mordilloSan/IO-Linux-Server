// src/types/theme.d.ts

import "@mui/material/styles";
import "@emotion/react";
import { Theme as MuiTheme } from "@mui/material/styles";

// Shared custom extension
interface CustomTheme {
  name: string;
  header: {
    color: string;
    background: string;
    search: {
      color: string;
    };
    indicator: {
      background: string;
    };
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
      brand: {
        color: string;
      };
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
}

// 👉 Extend MUI theme
declare module "@mui/material/styles" {
  interface Theme extends CustomTheme {}
  interface ThemeOptions extends Partial<CustomTheme> {}
}

// 👉 Extend Emotion's theme (used in styled components)
declare module "@emotion/react" {
  /* eslint-disable @typescript-eslint/no-empty-interface */
  export interface Theme extends MuiTheme {}
  /* eslint-disable @typescript-eslint/no-empty-interface */
}

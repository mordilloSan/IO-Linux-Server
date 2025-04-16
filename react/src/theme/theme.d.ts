import "@emotion/react";
import { Theme as MuiTheme } from "@mui/material/styles";

interface CustomTheme extends MuiTheme {
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

declare module "@emotion/react" {
  export interface Theme extends CustomTheme {}
}

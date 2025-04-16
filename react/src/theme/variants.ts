import { grey } from "@mui/material/colors";
import { THEMES } from "@/constants";
import { VariantType } from "@/types/theme";

const customBlue = {
  500: "#4782da",
  600: "#407ad6",
  700: "#376fd0",
};

export const variants: VariantType[] = [
  {
    name: THEMES.DEFAULT,
    palette: {
      mode: "light",
      primary: {
        main: customBlue[700],
        contrastText: "#FFF",
      },
      secondary: {
        main: customBlue[500],
        contrastText: "#FFF",
      },
      background: {
        default: "#F7F9FC",
        paper: "#FFF",
      },
      text: {
        primary: grey[900],
        secondary: grey[600],
      },
    },
    header: {
      color: grey[500],
      background: "#FFF",
      search: { color: grey[800] },
      indicator: { background: customBlue[600] },
    },
    footer: {
      color: grey[500],
      background: "#FFF",
    },
    sidebar: {
      color: grey[200],
      background: "#233044",
      header: {
        color: grey[200],
        background: "#233044",
        brand: { color: customBlue[500] },
      },
      footer: {
        color: grey[200],
        background: "#1E2A38",
      },
      badge: {
        color: "#FFF",
        background: customBlue[500],
      },
    },
  },

  {
    name: THEMES.DARK,
    palette: {
      mode: "dark",
      primary: {
        main: customBlue[600],
        contrastText: "#FFF",
      },
      secondary: {
        main: customBlue[500],
        contrastText: "#FFF",
      },
      background: {
        default: "#1B2635",
        paper: "#233044",
      },
      text: {
        primary: "rgba(255, 255, 255, 0.95)",
        secondary: "rgba(255, 255, 255, 0.5)",
      },
    },
    header: {
      color: grey[300],
      background: "#1B2635",
      search: { color: grey[200] },
      indicator: { background: customBlue[500] },
    },
    footer: {
      color: grey[300],
      background: "#1B2635",
    },
    sidebar: {
      color: grey[200],
      background: "#1E2A38",
      header: {
        color: grey[200],
        background: "#233044",
        brand: { color: customBlue[500] },
      },
      footer: {
        color: grey[200],
        background: "#1B2635",
      },
      badge: {
        color: "#FFF",
        background: customBlue[500],
      },
    },
  },
];

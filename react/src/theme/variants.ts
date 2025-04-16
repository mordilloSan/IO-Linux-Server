import { grey } from "@mui/material/colors";
import { THEMES } from "@/constants";

const customBlue = {
  50: "#e9f0fb",
  100: "#c8daf4",
  200: "#a3c1ed",
  300: "#7ea8e5",
  400: "#6395e0",
  500: "#4782da",
  600: "#407ad6",
  700: "#376fd0",
  800: "#2f65cb",
  900: "#2052c2",
};

const defaultVariant = {
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
};

const darkVariant = {
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
};

const variants = [defaultVariant, darkVariant];

export default variants;

export type VariantType = {
  name: string;
  palette: {
    mode: "light" | "dark";
    primary: MainContrastTextType;
    secondary: MainContrastTextType;
    background: {
      default: string;
      paper: string;
    };
    text?: {
      primary?: string;
      secondary?: string;
    };
  };
};

type MainContrastTextType = {
  main: string;
  contrastText: string;
};

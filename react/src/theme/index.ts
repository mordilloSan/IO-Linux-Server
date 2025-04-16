import "@mui/lab/themeAugmentation";
import { createTheme as createMuiTheme } from "@mui/material/styles";
import { variants } from "@/theme/variants";
import typography from "@/theme/typography";
import breakpoints from "@/theme/breakpoints";
import components from "@/theme/components";
import shadows from "@/theme/shadows";

const createTheme = (name: string) => {
  const themeConfig = variants.find((v) => v.name === name) ?? variants[0];

  // First object: MUI ThemeOptions
  const baseTheme = createMuiTheme({
    spacing: 4,
    breakpoints,
    components,
    typography,
    shadows,
    palette: themeConfig.palette,
  });

  // Merge in custom props
  return {
    ...baseTheme,
    name: themeConfig.name,
    header: themeConfig.header,
    footer: themeConfig.footer,
    sidebar: themeConfig.sidebar,
  };
};

export default createTheme;

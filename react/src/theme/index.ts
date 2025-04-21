import "@mui/lab/themeAugmentation";

import { createTheme as createMuiTheme } from "@mui/material/styles";
import variants from "@/theme/variants";
import typography from "@/theme/typography";
import breakpoints from "@/theme/breakpoints";
import components from "@/theme/components";
import shadows from "@/theme/shadows";

const createTheme = (name: string, primaryColor?: string) => {
  let themeConfig = variants.find((variant) => variant.name === name);

  if (!themeConfig) {
    console.warn(new Error(`The theme ${name} is not valid`));
    themeConfig = variants[0];
  }

  const palette = {
    ...themeConfig.palette,
    primary: {
      ...themeConfig.palette.primary,
      ...(primaryColor && { main: primaryColor }),
    },
  };

  return createMuiTheme(
    {
      spacing: 4,
      breakpoints: breakpoints,
      // @ts-expect-error Incompatible types
      components: components,
      typography: typography,
      shadows: shadows,
      palette,
    },
    {
      name: themeConfig.name,
      header: themeConfig.header,
      footer: themeConfig.footer,
      sidebar: themeConfig.sidebar,
    }
  );
};

export default createTheme;

import React, { useMemo } from "react";
import { useRoutes } from "react-router-dom";
import { CacheProvider } from "@emotion/react";
import { ThemeProvider as MuiThemeProvider } from "@mui/material/styles";

import createTheme from "./theme";
import routes from "./routes";

import useTheme from "@/hooks/useTheme";

import createEmotionCache from "@/utils/createEmotionCache";
import ReactQueryProvider from "./utils/ReactQueryProvider";

import { AuthProvider } from "./contexts/AuthContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";

const clientSideEmotionCache = createEmotionCache();

function App({ emotionCache = clientSideEmotionCache }) {
  const content = useRoutes(routes);
  const { theme: themeName, primaryColor } = useTheme();
  const theme = useMemo(
    () => createTheme(themeName, primaryColor),
    [themeName, primaryColor]
  );

  return (
    <CacheProvider value={emotionCache}>
      <MuiThemeProvider theme={theme}>
        <ReactQueryProvider>
          <AuthProvider>
            <WebSocketProvider>{content}</WebSocketProvider>
          </AuthProvider>
        </ReactQueryProvider>
      </MuiThemeProvider>
    </CacheProvider>
  );
}

export default App;

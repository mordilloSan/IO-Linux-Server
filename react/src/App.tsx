import React from "react";
import { useRoutes } from "react-router-dom";

import { CacheProvider } from "@emotion/react";

import { ThemeProvider as MuiThemeProvider } from "@mui/material/styles";

import createTheme from "./theme";
import routes from "./routes";

import useTheme from "@/hooks/useTheme";
import createEmotionCache from "@/utils/createEmotionCache";

import { AuthProvider } from "./contexts/AuthContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";

const clientSideEmotionCache = createEmotionCache();

function App({ emotionCache = clientSideEmotionCache }) {
  const content = useRoutes(routes);

  const { theme } = useTheme();

  return (
    <CacheProvider value={emotionCache}>
      <MuiThemeProvider theme={createTheme(theme)}>
        <AuthProvider>
          <WebSocketProvider>{content}</WebSocketProvider>
        </AuthProvider>
      </MuiThemeProvider>
    </CacheProvider>
  );
}

export default App;

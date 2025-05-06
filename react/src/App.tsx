import React, { useMemo } from "react";
import { useRoutes } from "react-router-dom";
import { CacheProvider } from "@emotion/react";
import { ThemeProvider as MuiThemeProvider } from "@mui/material/styles";

import createTheme from "./theme";
import routes from "./routes";

import useTheme from "@/hooks/useAppTheme";

import createEmotionCache from "@/utils/createEmotionCache";
import ReactQueryProvider from "./utils/ReactQueryProvider";

import { AuthProvider } from "./contexts/AuthContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";
import { SidebarProvider } from "@/contexts/SidebarContext";
import { Toaster } from "sonner";

const clientSideEmotionCache = createEmotionCache();

function App({ emotionCache = clientSideEmotionCache }) {
  const content = useRoutes(routes);
  const { theme: themeName, primaryColor } = useTheme();
  const theme = useMemo(
    () => createTheme(themeName, primaryColor),
    [themeName, primaryColor],
  );

  return (
    <CacheProvider value={emotionCache}>
      <MuiThemeProvider theme={theme}>
        <ReactQueryProvider>
          <AuthProvider>
            <WebSocketProvider>
              <SidebarProvider>{content}</SidebarProvider>
            </WebSocketProvider>
          </AuthProvider>
        </ReactQueryProvider>
        <Toaster richColors position="top-right" />
      </MuiThemeProvider>
    </CacheProvider>
  );
}

export default App;

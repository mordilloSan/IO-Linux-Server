import React, {
  createContext,
  useEffect,
  useReducer,
  useCallback,
  useMemo,
} from "react";

import {
  AuthContextType,
  AuthState,
  AuthActions,
  AuthProviderProps,
  AUTH_ACTIONS,
  AuthUser,
} from "@/types/auth";
import axios from "@/utils/axios";

const initialState: AuthState = {
  isAuthenticated: false,
  isInitialized: false,
  user: null,
};

const reducer = (state: AuthState, action: AuthActions): AuthState => {
  switch (action.type) {
    case AUTH_ACTIONS.INITIALIZE_START:
      return { ...state, isInitialized: false };
    case AUTH_ACTIONS.INITIALIZE_SUCCESS:
      return {
        ...state,
        isInitialized: true,
        isAuthenticated: true,
        user: action.payload.user,
      };
    case AUTH_ACTIONS.INITIALIZE_FAILURE:
      return {
        ...state,
        isInitialized: true,
        isAuthenticated: false,
        user: null,
      };
    case AUTH_ACTIONS.SIGN_IN:
      return { ...state, isAuthenticated: true, user: action.payload.user };
    case AUTH_ACTIONS.SIGN_OUT:
      return { ...state, isAuthenticated: false, user: null };
    default: {
      const exhaustiveCheck: never = action;
      void exhaustiveCheck; // TypeScript exhaustiveness check
      return state;
    }
  }
};

const AuthContext = createContext<AuthContextType | null>(null);
AuthContext.displayName = "AuthContext";

function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(reducer, initialState);

  // Memoize fetchUser so signIn and initialize can depend on it
  const fetchUser = useCallback(async (): Promise<AuthUser> => {
    const { data } = await axios.get<{ user: AuthUser }>("/auth/me");
    return data.user;
  }, []);

  const initialize = useCallback(async () => {
    dispatch({ type: AUTH_ACTIONS.INITIALIZE_START });
    try {
      const user = await fetchUser();
      dispatch({ type: AUTH_ACTIONS.INITIALIZE_SUCCESS, payload: { user } });
    } catch {
      dispatch({ type: AUTH_ACTIONS.INITIALIZE_FAILURE });
    }
  }, [fetchUser]);

  useEffect(() => {
    initialize();
  }, [initialize]);

  useEffect(() => {
    // Only run after initial auth check is complete
    if (!state.isInitialized) return;

    const handleVisibilityOrFocus = () => {
      if (document.visibilityState === "visible") {
        initialize();
      }
    };

    window.addEventListener("visibilitychange", handleVisibilityOrFocus);
    window.addEventListener("focus", handleVisibilityOrFocus);

    return () => {
      window.removeEventListener("visibilitychange", handleVisibilityOrFocus);
      window.removeEventListener("focus", handleVisibilityOrFocus);
    };
  }, [initialize, state.isInitialized]);

  useEffect(() => {
    const handleStorage = (e: StorageEvent) => {
      if (e.key === "logout") {
        dispatch({ type: AUTH_ACTIONS.SIGN_OUT });
        window.location.href = "/sign-in";
      }
    };
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  // Memoized signIn and signOut
  const signIn = useCallback(
    async (username: string, password: string) => {
      await axios.post("/auth/login", { username, password });
      const user = await fetchUser();
      dispatch({ type: AUTH_ACTIONS.SIGN_IN, payload: { user } });
    },
    [fetchUser],
  );

  const signOut = useCallback(async () => {
    await axios.get("/auth/logout");
    localStorage.setItem("logout", Date.now().toString()); // Broadcast logout
    dispatch({ type: AUTH_ACTIONS.SIGN_OUT });
  }, []);

  // Memoize context value for optimal render performance
  const contextValue = useMemo(
    () => ({
      ...state,
      method: "session" as const,
      signIn,
      signOut,
    }),
    [state, signIn, signOut],
  );

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}

export { AuthContext, AuthProvider };

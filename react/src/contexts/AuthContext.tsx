import React, {
  createContext,
  useEffect,
  useReducer,
  useCallback,
} from "react";
import { useLocation } from "react-router-dom";

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
      void exhaustiveCheck; // mark as intentionally unused
      return state;
    }
  }
};

const AuthContext = createContext<AuthContextType | null>(null);
AuthContext.displayName = "AuthContext";

function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(reducer, initialState);
  const location = useLocation();

  const fetchUser = async (): Promise<AuthUser> => {
    const { data } = await axios.get<{ user: AuthUser }>("/auth/me");
    return data.user;
  };

  const initialize = useCallback(async () => {
    dispatch({ type: AUTH_ACTIONS.INITIALIZE_START });
    try {
      const user = await fetchUser();
      dispatch({ type: AUTH_ACTIONS.INITIALIZE_SUCCESS, payload: { user } });
    } catch {
      dispatch({ type: AUTH_ACTIONS.INITIALIZE_FAILURE });
    }
  }, []);

  useEffect(() => {
    initialize();
  }, [location.pathname, initialize]);

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

  const signIn = async (username: string, password: string) => {
    await axios.post("/auth/login", { username, password });
    const user = await fetchUser();
    dispatch({ type: AUTH_ACTIONS.SIGN_IN, payload: { user } });
  };

  const signOut = async () => {
    await axios.get("/auth/logout");
    localStorage.setItem("logout", Date.now().toString()); // Broadcast logout
    dispatch({ type: AUTH_ACTIONS.SIGN_OUT });
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
        method: "session",
        signIn,
        signOut,
      }}>
      {children}
    </AuthContext.Provider>
  );
}

export { AuthContext, AuthProvider };

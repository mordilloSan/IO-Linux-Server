import React, {
  createContext,
  ReactNode,
  useEffect,
  useReducer,
  useCallback,
} from "react";
import { useLocation } from "react-router-dom";
import axios from "@/utils/axios";
import { AuthContextType, ActionMap, AuthState, AuthUser } from "@/types/auth";
import { useContext } from "react";
import { WebSocketContext } from "@/contexts/WebSocketContext";

const API_URL = import.meta.env.VITE_API_URL;

const INITIALIZE = "INITIALIZE";
const SIGN_IN = "SIGN_IN";
const SIGN_OUT = "SIGN_OUT";

type AuthActionTypes = {
  [INITIALIZE]: { isAuthenticated: boolean; user: AuthUser | null };
  [SIGN_IN]: { user: AuthUser };
  [SIGN_OUT]: undefined;
};

const initialState: AuthState = {
  isAuthenticated: false,
  isInitialized: false,
  user: null,
};

const reducer = (
  state: AuthState,
  action: ActionMap<AuthActionTypes>[keyof ActionMap<AuthActionTypes>]
): AuthState => {
  switch (action.type) {
    case INITIALIZE:
      return { ...state, isInitialized: true, ...action.payload };
    case SIGN_IN:
      return { ...state, isAuthenticated: true, user: action.payload.user };
    case SIGN_OUT:
      return { ...state, isAuthenticated: false, user: null };
    default:
      return state;
  }
};

const AuthContext = createContext<AuthContextType | null>(null);

function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState);
  const location = useLocation();
  const { disconnect } = useContext(WebSocketContext);
  const fetchUser = async () => {
    const response = await axios.get(`${API_URL}/auth/me`);
    return response.data.user;
  };

  const initialize = useCallback(async () => {
    try {
      const user = await fetchUser();
      dispatch({
        type: INITIALIZE,
        payload: { isAuthenticated: true, user },
      });
    } catch {
      dispatch({ type: SIGN_OUT });
    }
  }, []);

  useEffect(() => {
    initialize(); // always re-check session on route change
  }, [location.pathname, initialize]);

  useEffect(() => {
    if (!state.isAuthenticated && state.isInitialized) {
      initialize();
    }
  }, [
    location.pathname,
    state.isAuthenticated,
    state.isInitialized,
    initialize,
  ]);

  const signIn = async (username: string, password: string) => {
    await axios.post(`${API_URL}/auth/login`, { username, password });
    const user = await fetchUser();
    dispatch({ type: SIGN_IN, payload: { user } });
  };

  const signOut = async () => {
    await axios.get(`${API_URL}/auth/logout`);
    disconnect(); // 👈 explicitly close the socket
    dispatch({ type: SIGN_OUT });
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
        method: "session",
        signIn,
        signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export { AuthContext, AuthProvider };

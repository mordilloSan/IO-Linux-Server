import React, {
  createContext,
  ReactNode,
  useEffect,
  useReducer,
  useCallback,
} from "react";
import { useLocation } from "react-router-dom";
import { toast } from "sonner";

import { AuthContextType, ActionMap, AuthState, AuthUser } from "@/types/auth";
import axios from "@/utils/axios";
import { getErrorMessage } from "@/utils/getErrorMessage";

const INITIALIZE_START = "INITIALIZE_START";
const INITIALIZE_SUCCESS = "INITIALIZE_SUCCESS";
const INITIALIZE_FAILURE = "INITIALIZE_FAILURE";
const SIGN_IN = "SIGN_IN";
const SIGN_OUT = "SIGN_OUT";

type AuthActionTypes = {
  [INITIALIZE_START]: undefined;
  [INITIALIZE_SUCCESS]: { user: AuthUser };
  [INITIALIZE_FAILURE]: undefined;
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
  action: ActionMap<AuthActionTypes>[keyof ActionMap<AuthActionTypes>],
): AuthState => {
  switch (action.type) {
    case INITIALIZE_START:
      return { ...state, isInitialized: false };
    case INITIALIZE_SUCCESS:
      return {
        ...state,
        isInitialized: true,
        isAuthenticated: true,
        user: action.payload.user,
      };
    case INITIALIZE_FAILURE:
      return {
        ...state,
        isInitialized: true,
        isAuthenticated: false,
        user: null,
      };
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

  const fetchUser = async () => {
    const response = await axios.get("/auth/me");
    return response.data.user;
  };

  const initialize = useCallback(async () => {
    dispatch({ type: INITIALIZE_START });
    try {
      const user = await fetchUser();
      dispatch({ type: INITIALIZE_SUCCESS, payload: { user } });
    } catch {
      dispatch({ type: INITIALIZE_FAILURE });
    }
  }, []);

  useEffect(() => {
    initialize(); // always re-check session on route change
  }, [location.pathname, initialize]);

  const signIn = async (username: string, password: string) => {
    try {
      await axios.post("/auth/login", { username, password });
      const user = await fetchUser();
      dispatch({ type: SIGN_IN, payload: { user } });
    } catch (err) {
      toast.error(getErrorMessage(err));
      throw err;
    }
  };

  const signOut = async () => {
    try {
      await axios.get("/auth/logout");
      dispatch({ type: SIGN_OUT });
    } catch (err) {
      toast.error(getErrorMessage(err));
      throw err;
    }
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

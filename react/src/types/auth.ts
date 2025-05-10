export type ActionMap<M extends { [index: string]: any }> = {
  [Key in keyof M]: M[Key] extends undefined
    ? { type: Key }
    : { type: Key; payload: M[Key] };
};

export type AuthUser = {
  id: string;
  name: string;
};

export type AuthState = {
  isAuthenticated: boolean;
  isInitialized: boolean;
  user: AuthUser | null;
};

export type AuthContextType = {
  isAuthenticated: boolean;
  isInitialized: boolean;
  user: AuthUser | null;
  method: "session";
  signIn: (username: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
};

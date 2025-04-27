import React from "react";
import { Navigate } from "react-router-dom";

import useAuth from "@/hooks/useAuth";

interface AuthGuardType {
  children: React.ReactNode;
}

// For routes that can only be accessed by authenticated users
function AuthGuard({ children }: AuthGuardType) {
  const { isAuthenticated, isInitialized } = useAuth();

  if (!isInitialized) {
    return <Navigate to="/sign-in" replace />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/sign-in" replace />;
  }

  return <>{children}</>;
}

export default AuthGuard;

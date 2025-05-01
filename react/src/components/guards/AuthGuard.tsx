import React from "react";
import { Navigate, useLocation } from "react-router-dom";

import useAuth from "@/hooks/useAuth";
import PageLoader from "../PageLoader";

interface AuthGuardType {
  children: React.ReactNode;
}

// For routes that must be accessed only by authenticated users
function AuthGuard({ children }: AuthGuardType) {
  const { isAuthenticated, isInitialized } = useAuth();
  const location = useLocation();

  const redirectPath = `/sign-in?redirect=${encodeURIComponent(
    location.pathname + location.search
  )}`;

  if (!isInitialized) {
    return <Navigate to={redirectPath} replace />;
  }

  if (!isAuthenticated) {
    return <Navigate to={redirectPath} replace />;
  }

  return <>{children}</>;
}

export default AuthGuard;

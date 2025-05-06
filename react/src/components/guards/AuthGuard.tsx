import React from "react";
import { Navigate, useLocation } from "react-router-dom";
import useAuth from "@/hooks/useAuth";

interface AuthGuardType {
  children: React.ReactNode;
}

function AuthGuard({ children }: AuthGuardType) {
  const { isAuthenticated, isInitialized } = useAuth();
  const location = useLocation();

  if (isInitialized && !isAuthenticated) {
    const redirectPath = `/sign-in?redirect=${encodeURIComponent(
      location.pathname + location.search,
    )}`;
    return <Navigate to={redirectPath} replace />;
  }

  return <>{children}</>;
}

export default AuthGuard;

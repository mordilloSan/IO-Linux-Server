import React from "react";
import { Navigate, useSearchParams } from "react-router-dom";
import useAuth from "@/hooks/useAuth";
import PageLoader from "../PageLoader";

interface GuestGuardType {
  children: React.ReactNode;
}

// For routes that can only be accessed by unauthenticated users
function GuestGuard({ children }: GuestGuardType) {
  const { isAuthenticated, isInitialized } = useAuth();
  const [searchParams] = useSearchParams();
  const redirect = searchParams.get("redirect") || "/";

  if (isInitialized && isAuthenticated) {
    return <Navigate to={redirect} replace />;
  }

  return <>{children}</>;
}

export default GuestGuard;

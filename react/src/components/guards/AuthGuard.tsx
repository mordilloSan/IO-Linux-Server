import React, { PropsWithChildren } from "react";
import { Navigate, useLocation } from "react-router-dom";

import PageLoader from "../loaders/PageLoader";

import useAuth from "@/hooks/useAuth";

/**
 * Protects routes from unauthenticated access.
 *
 * Renders a loading component until auth is initialized,
 * and redirects to `/sign-in?redirect=...` if not authenticated.
 *
 * @example
 * ```tsx
 * <Route
 *   path="/admin"
 *   element={
 *     <AuthGuard>
 *       <AdminPage />
 *     </AuthGuard>
 *   }
 * />
 * ```
 */
export const AuthGuard: React.FC<PropsWithChildren> = ({ children }) => {
  const { isAuthenticated, isInitialized } = useAuth();
  const location = useLocation();

  if (!isInitialized) {
    return <PageLoader />;
  }

  if (!isAuthenticated) {
    const redirectPath = `/sign-in?redirect=${encodeURIComponent(
      location.pathname + location.search,
    )}`;
    return <Navigate to={redirectPath} replace />;
  }

  return <>{children}</>;
};

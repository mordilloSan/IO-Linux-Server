// React
import React, { lazy } from "react";

// Guards & Layouts
import { AuthGuard } from "@/components/guards/AuthGuard";
import { GuestGuard } from "@/components/guards/GuestGuard";
import AuthLayout from "@/layouts/Auth";
import MainLayout from "@/layouts/Main";
import Default from "@/pages/dashboard/home";

// Helper: Loadable wrapper
const Loadable = (Component: React.LazyExoticComponent<any>, name: string) => {
  const Wrapped = (props: any) => <Component {...props} />;
  Wrapped.displayName = `Loadable(${name})`;
  return Wrapped;
};

const lazyLoad = (
  factory: () => Promise<{ default: React.ComponentType<any> }>,
  name: string
) => Loadable(lazy(factory), name);

// Lazy-loaded pages
const SignIn = lazyLoad(() => import("@/pages/auth/SignIn"), "SignIn");
const Page404 = lazyLoad(() => import("@/pages/auth/Page404"), "Page404");
const Updates = lazyLoad(() => import("@/pages/dashboard/updates"), "Updates");
const Docker = lazyLoad(() => import("@/pages/dashboard/docker"), "Docker");
const Services = lazyLoad(() => import("@/pages/dashboard/services"), "Docker");

const routes = [
  {
    path: "/",
    element: (
      <AuthGuard>
        <MainLayout />
      </AuthGuard>
    ),
    children: [
      { path: "", element: <Default /> },
      { path: "updates", element: <Updates /> },
      { path: "docker", element: <Docker /> },
      { path: "services", element: <Services /> },
    ],
  },
  {
    path: "*",
    element: <AuthLayout />,
    children: [
      {
        path: "sign-in",
        element: (
          <GuestGuard>
            <SignIn />
          </GuestGuard>
        ),
      },
      { path: "*", element: <Page404 /> },
    ],
  },
];

export default routes;

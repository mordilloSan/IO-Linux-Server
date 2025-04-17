// React
import React, { lazy, Suspense } from "react";

// Layouts
import AuthLayout from "@/layouts/Auth";
import ErrorLayout from "@/layouts/Error";
import DashboardLayout from "@/layouts/Dashboard";

// Guards
import AuthGuard from "@/components/guards/AuthGuard";
import GuestGuard from "./components/guards/GuestGuard";
import Loader from "./components/Loader";

const Loadable = (Component: React.LazyExoticComponent<any>, name: string) => {
  const Wrapped = (props: any) => (
    <Suspense fallback={<Loader />}>
      <Component {...props} />
    </Suspense>
  );

  Wrapped.displayName = `Loadable(${name})`;
  return Wrapped;
};

// Lazy-loaded components
const SignIn = Loadable(
  lazy(() => import("@/pages/auth/SignIn")),
  "SignIn"
);
const Page404 = Loadable(
  lazy(() => import("@/pages/auth/Page404")),
  "Page404"
);
const Blank = Loadable(
  lazy(() => import("@/pages/dashboards/Blank")),
  "Blank"
);
const Orders = Loadable(
  lazy(() => import("@/pages/dashboards/Orders")),
  "Orders"
);
const Settings = Loadable(
  lazy(() => import("@/pages/dashboards/Settings")),
  "Settings"
);
const Default = Loadable(
  lazy(() => import("@/pages/dashboards")),
  "Default"
);

const routes = [
  {
    path: "/",
    element: (
      <AuthGuard>
        <DashboardLayout />
      </AuthGuard>
    ),
    children: [
      { path: "", element: <Default /> },
      { path: "blank", element: <Blank /> },
      { path: "settings", element: <Settings /> },
      { path: "orders", element: <Orders /> },
    ],
  },
  {
    path: "auth",
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
    ],
  },
  {
    path: "*",
    element: <ErrorLayout />,
    children: [{ path: "*", element: <Page404 /> }],
  },
];

export default routes;

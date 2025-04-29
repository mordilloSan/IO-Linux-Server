// React
import React, { lazy, Suspense } from "react";

// Layouts
import AuthLayout from "@/layouts/Auth";
import MainLayout from "@/layouts/Main";

// Guards
import AuthGuard from "@/components/guards/AuthGuard";
import GuestGuard from "./components/guards/GuestGuard";
import PageLoader from "./components/PageLoader";
import Docker from "./pages/dashboard/docker";

const Loadable = (Component: React.LazyExoticComponent<any>, name: string) => {
  const Wrapped = (props: any) => (
    <Suspense fallback={<PageLoader />}>
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
const Default = Loadable(
  lazy(() => import("@/pages/dashboard/home")),
  "Default"
);

const Updates = Loadable(
  lazy(() => import("@/pages/dashboard/updates")),
  "Default"
);

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
    ],
  },

  {
    path: "*",
    element: <AuthLayout />,
    children: [
      { path: "*", element: <Page404 /> },
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
];

export default routes;

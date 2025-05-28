import React, { Suspense } from "react";
import NetworkInterfaceList from "./NetworkInterfaceList";
import ComponentLoader from "@/components/loaders/ComponentLoader";

const NetworkPage: React.FC = () => (
  <Suspense fallback={<ComponentLoader />}>
    <NetworkInterfaceList />
  </Suspense>
);

export default NetworkPage;

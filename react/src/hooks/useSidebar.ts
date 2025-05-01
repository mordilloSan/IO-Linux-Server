import { SidebarContext, SidebarContextType } from "@/contexts/SidebarContext";
import { useContext } from "react";

const useSidebar = (): SidebarContextType => {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within a SidebarProvider");
  }
  return context;
};

export default useSidebar;

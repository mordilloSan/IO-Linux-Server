import { useContext } from "react";
import { AuthContext } from "@/contexts/AuthContext";
import { AuthContextType } from "@/types/auth";

const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);

  if (!context)
    throw new Error("AuthContext must be placed within AuthProvider");

  return context;
};

export default useAuth;

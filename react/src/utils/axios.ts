// src/utils/axios.ts
import axios, { AxiosError } from "axios";

const axiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
});

axiosInstance.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response) {
      const status = error.response.status;

      if (
        status === 401 &&
        !error.config?.url?.includes("/auth/me") // ⛔️ Don't redirect if the request was to /auth/me
      ) {
        const redirectPath = window.location.pathname + window.location.search;
        window.location.href = `/auth/sign-in?redirect=${encodeURIComponent(
          redirectPath
        )}`;
        return;
      }

      if (status === 500) {
        window.location.href = "/error/500";
        return;
      }

      return Promise.reject(error);
    }

    // Low-level network or timeout error
    console.error("Network error:", error.message);
    return Promise.reject(error);
  }
);

export default axiosInstance;

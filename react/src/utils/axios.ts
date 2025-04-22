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

      if (status === 401) {
        // Redirect to signin on Unauthorized
        window.location.href = "/auth/signin";
        return; // prevent further handling
      }

      if (status === 500) {
        // Redirect to generic error page
        window.location.href = "/error/500";
        return;
      }

      // Reject for React Query to handle (e.g. toast)
      return Promise.reject(error);
    }

    // Low-level network or timeout error
    console.error("Network error:", error.message);
    return Promise.reject(error);
  }
);

export default axiosInstance;

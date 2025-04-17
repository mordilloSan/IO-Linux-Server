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

      if (status === 500) {
        // Only redirect — let React Query handle the rest (toast, retry, etc.)
        window.location.href = "/error/500";
      }

      // Reject full error for React Query to catch and show via toast
      return Promise.reject(error);
    }

    // Optional: toast only for low-level network failure (React Query won't catch this)
    console.error("Network error:", error.message);
    return Promise.reject(error);
  }
);

export default axiosInstance;

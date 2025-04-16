import axios from "axios";

const axiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
});

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      if (error.response.status === 500) {
        window.location.href = "/error/500";
      }

      return Promise.reject(error.response.data);
    }

    return Promise.reject("Network error, please try again later.");
  }
);

export default axiosInstance;

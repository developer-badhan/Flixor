import api from './api';

export interface UserProfile {
  id: string;
  username: string;
  email: string;
  profile_picture: string;
  is_verified: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * Service for user-related API calls (fetch user data, update profile, upload picture, send/verify OTP, get analytics)
 * uses axios instance for API calls
 * 
 * @returns {Object} - Object containing methods for user-related API calls
 * @property {Function} getMe - Fetches user data
 * @property {Function} updateProfile - Updates user profile
 * @property {Function} uploadProfilePicture - Uploads user profile picture
 * @property {Function} sendOtp - Sends OTP to user's email
 * @property {Function} verifyOtp - Verifies user's OTP
 * @returns {Promise<UserProfile>} - Promise that resolves to user profile object
 */

export const userService = {
  getMe: async (): Promise<UserProfile> => {
    const response = await api.get('/user/me');
    return response.data;
  },

  updateProfile: async (data: { username?: string; password?: string }): Promise<UserProfile> => {
    const response = await api.patch('/user/profile', data);
    return response.data;
  },

  uploadProfilePicture: async (file: File): Promise<UserProfile> => {
    const formData = new FormData();
    formData.append('picture', file);
    const response = await api.post('/user/profile-picture', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  sendOtp: async (): Promise<{ success: boolean; message: string }> => {
    const response = await api.post('/user/send-otp');
    return response.data; // Note: Our api interceptor unwraps success/message for some requests, let's just return what axios resolves
  },

  verifyOtp: async (otp: string): Promise<{ success: boolean; message: string }> => {
    const response = await api.post('/user/verify-otp', { otp });
    return response.data;
  },
};

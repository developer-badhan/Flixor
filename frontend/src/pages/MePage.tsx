import React, { useState, useRef, useEffect } from 'react';
import { useUserContext } from '../context/UserContext';
import { userService } from '../services/userService';
import { BadgeCheck, XCircle, Camera, Loader2, Save } from 'lucide-react';
import { Helmet } from 'react-helmet-async';

/**
 * MePage is a React component that displays the user's profile information.
 * It allows users to view and edit their profile information, including username, password, and profile picture.
 * It also allows users to verify their email address by sending and verifying OTP.
 * 
 * @returns {JSX.Element} - Returns the MePage component
 */

const MePage: React.FC = () => {
  const { user, loading: userLoading, refreshUser } = useUserContext();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [updatingProfile, setUpdatingProfile] = useState(false);
  const [profileMsg, setProfileMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [otpSent, setOtpSent] = useState(false);
  const [otp, setOtp] = useState('');
  const [verifyingOtp, setVerifyingOtp] = useState(false);
  const [otpMsg, setOtpMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    if (user) {
      setUsername(user.username || '');
    }
  }, [user]);

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setUpdatingProfile(true);
    setProfileMsg(null);
    try {
      const data: any = {};
      if (username !== user?.username) data.username = username;
      if (password) data.password = password;

      if (Object.keys(data).length === 0) {
        setProfileMsg({ type: 'success', text: 'No changes made.' });
        setUpdatingProfile(false);
        return;
      }

      await userService.updateProfile(data);
      setProfileMsg({ type: 'success', text: 'Profile updated successfully!' });
      setPassword('');
      await refreshUser();
    } catch (error: any) {
      setProfileMsg({ type: 'error', text: error.response?.data?.error || 'Failed to update profile.' });
    } finally {
      setUpdatingProfile(false);
    }
  };

  /**
   * Handles the avatar upload process.
   * @param {React.ChangeEvent<HTMLInputElement>} e - The file input change event.
   * @returns {Promise<void>} - Returns a promise that resolves when the avatar upload is complete.
   */
  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return;
    
    setUploadingAvatar(true);
    try {
      await userService.uploadProfilePicture(e.target.files[0]);
      await refreshUser();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to upload image.');
    } finally {
      setUploadingAvatar(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  /**
   * Sends OTP to the user's email.
   * @returns {Promise<void>} - Returns a promise that resolves when the OTP is sent successfully.
   */
  const handleSendOtp = async () => {
    setOtpMsg(null);
    try {
      await userService.sendOtp();
      setOtpSent(true);
      setOtpMsg({ type: 'success', text: 'OTP sent to your email.' });
    } catch (error: any) {
      setOtpMsg({ type: 'error', text: error.response?.data?.error || 'Failed to send OTP.' });
    }
  };

  /**
   * Verifies the user's OTP.
   * @returns {Promise<void>} - Returns a promise that resolves when the OTP is verified successfully.
   */
  const handleVerifyOtp = async () => {
    if (otp.length !== 6) {
      setOtpMsg({ type: 'error', text: 'OTP must be 6 digits.' });
      return;
    }
    setVerifyingOtp(true);
    setOtpMsg(null);
    try {
      await userService.verifyOtp(otp);
      setOtpMsg({ type: 'success', text: 'Email verified successfully!' });
      await refreshUser();
    } catch (error: any) {
      setOtpMsg({ type: 'error', text: error.response?.data?.error || 'Invalid or expired OTP.' });
    } finally {
      setVerifyingOtp(false);
    }
  };

  if (userLoading && !user) {
    return (
      <div className="flex h-screen items-center justify-center bg-flixor-dark">
        <Loader2 className="animate-spin text-flixor-red w-12 h-12" />
      </div>
    );
  }

  if (!user) return null;

  return (
    <>
      <Helmet>
        <title>My Profile - Flixor</title>
      </Helmet>

      <div className="pt-24 pb-12 px-4 md:px-8 min-h-screen bg-flixor-dark flex justify-center items-start">
        <div className="w-full max-w-4xl grid grid-cols-1 md:grid-cols-3 gap-8">
          
          {/* Left Column: Avatar & Basic Info */}
          <div className="col-span-1 glass p-8 rounded-2xl border border-white/10 flex flex-col items-center shadow-xl">
            <div className="relative group mb-6">
              <div className="w-32 h-32 rounded-full overflow-hidden border-4 border-flixor-dark bg-black shadow-lg">
                <img
                  src={user.profile_picture || "https://cdn.iconscout.com/icon/premium/png-256-thumb/user-icon-svg-download-png-5358.png?f=webp&w=128"}
                  alt="Avatar"
                  className={`w-full h-full object-cover transition-opacity ${uploadingAvatar ? 'opacity-50' : 'group-hover:opacity-60'}`}
                  onError={(e) => {
                    (e.target as HTMLImageElement).src = "https://cdn.iconscout.com/icon/premium/png-256-thumb/user-icon-svg-download-png-5358.png?f=webp&w=128";
                  }}
                />
              </div>
              <button
                onClick={() => fileInputRef.current?.click()}
                disabled={uploadingAvatar}
                className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
              >
                {uploadingAvatar ? <Loader2 className="animate-spin text-white w-8 h-8" /> : <Camera className="text-white w-8 h-8" />}
              </button>
              <input
                type="file"
                accept="image/jpeg, image/png, image/webp"
                className="hidden"
                ref={fileInputRef}
                onChange={handleAvatarUpload}
              />
            </div>

            <h2 className="text-2xl font-bold text-white mb-1 flex items-center gap-2">
              {user.username}
              {user.is_verified ? (
                <span title="Verified Account"><BadgeCheck className="text-green-500 w-6 h-6" /></span>
              ) : (
                <span title="Unverified Account"><XCircle className="text-red-500 w-6 h-6" /></span>
              )}
            </h2>
            <p className="text-gray-400 text-sm mb-4">{user.email}</p>

            <div className="w-full text-left bg-black/40 rounded-lg p-4 text-sm text-gray-300">
              <p className="mb-2"><span className="text-gray-500">ID:</span> {user.id}</p>
              <p><span className="text-gray-500">Joined:</span> {new Date(user.created_at).toLocaleDateString()}</p>
            </div>
          </div>

          {/* Right Column: Edit Form & Verification */}
          <div className="col-span-1 md:col-span-2 space-y-8">
            
            {/* Edit Profile Section */}
            <div className="glass p-8 rounded-2xl border border-white/10 shadow-xl">
              <h3 className="text-xl font-bold text-white mb-6">Edit Profile</h3>
              
              <form onSubmit={handleProfileUpdate} className="space-y-5">
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-2">Username</label>
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="w-full bg-black/50 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-white/40 transition-colors"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-2">New Password (optional)</label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Leave blank to keep current password"
                    className="w-full bg-black/50 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-white/40 transition-colors"
                  />
                </div>

                {profileMsg && (
                  <div className={`p-3 rounded-lg text-sm ${profileMsg.type === 'success' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                    {profileMsg.text}
                  </div>
                )}

                <div className="pt-2">
                  <button
                    type="submit"
                    disabled={updatingProfile}
                    className="flex items-center gap-2 bg-white text-black font-semibold px-6 py-2.5 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-70"
                  >
                    {updatingProfile ? <Loader2 className="w-5 h-5 animate-spin" /> : <Save className="w-5 h-5" />}
                    Save Changes
                  </button>
                </div>
              </form>
            </div>

            {/* Verification Section */}
            {!user.is_verified && (
              <div className="glass p-8 rounded-2xl border border-flixor-red/30 shadow-xl">
                <h3 className="text-xl font-bold text-white mb-2 flex items-center gap-2">
                  <XCircle className="text-red-500 w-6 h-6" />
                  Email Verification
                </h3>
                <p className="text-gray-400 text-sm mb-6">Your email address is not verified. Please verify it to secure your account.</p>

                {otpMsg && (
                  <div className={`p-3 rounded-lg text-sm mb-4 ${otpMsg.type === 'success' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                    {otpMsg.text}
                  </div>
                )}

                {!otpSent ? (
                  <button
                    onClick={handleSendOtp}
                    className="bg-flixor-red text-white font-semibold px-6 py-2.5 rounded-lg hover:bg-red-700 transition-colors"
                  >
                    Send Verification Code
                  </button>
                ) : (
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-400 mb-2">Enter 6-digit OTP</label>
                      <input
                        type="text"
                        maxLength={6}
                        value={otp}
                        onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
                        className="w-full md:w-1/2 bg-black/50 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-white/40 tracking-widest"
                      />
                    </div>
                    <div className="flex gap-4">
                      <button
                        onClick={handleVerifyOtp}
                        disabled={verifyingOtp || otp.length !== 6}
                        className="bg-white text-black font-semibold px-6 py-2.5 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
                      >
                        {verifyingOtp ? 'Verifying...' : 'Verify Email'}
                      </button>
                      <button
                        onClick={handleSendOtp}
                        className="text-gray-400 hover:text-white px-4 py-2.5 text-sm transition-colors"
                      >
                        Resend Code
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

        </div>
      </div>
    </>
  );
};

export default MePage;

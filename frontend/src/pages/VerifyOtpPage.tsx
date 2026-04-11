import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { verifyOtp, sendOtp } from '../features/user/services/userService';

const VerifyOtpPage: React.FC = () => {
  const [otp, setOtp] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resendStatus, setResendStatus] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (otp.length !== 6) {
      setError('Please enter a 6-digit code.');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      await verifyOtp(otp);
      // Validated! Safe to enter home.
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to verify code. It might be invalid or expired.');
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setResendStatus('Resending...');
    setError(null);
    try {
      await sendOtp();
      setResendStatus('Code resent successfully!');
      setTimeout(() => setResendStatus(null), 3000);
    } catch (err: any) {
      setResendStatus(null);
      setError('Failed to resend code.');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-flixor-dark pt-20 px-4 relative">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-md w-full bg-black/80 border border-white/20 p-10 rounded-md shadow-2xl z-10 relative"
      >
        <h2 className="text-3xl font-bold mb-4 text-white text-center">Enter Code</h2>
        
        <p className="text-flixor-lightGray mb-8 text-center text-sm">
          Check your email! We've sent a 6-digit verification code to you.
        </p>

        {error && <div className="bg-flixor-red p-3 rounded mb-6 text-white text-sm text-center">{error}</div>}

        <form onSubmit={handleVerify} className="space-y-6">
          <div>
            <input 
              type="text" 
              maxLength={6}
              placeholder="000000" 
              className="w-full bg-[#333333] text-white px-4 py-4 rounded focus:outline-none focus:ring-2 focus:ring-white/30 text-center tracking-[1em] text-2xl font-bold"
              value={otp}
              onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))} // Only numbers
              required
            />
          </div>
          
          <button 
            type="submit" 
            className="w-full bg-flixor-red hover:bg-flixor-hover text-white font-bold py-3 rounded transition-colors"
            disabled={loading || otp.length !== 6}
          >
            {loading ? 'Verifying...' : 'Verify Now'}
          </button>
        </form>

        <div className="mt-8 text-[#b3b3b3] text-sm text-center">
          <p>Didn't receive the code?</p>
          <button 
            onClick={handleResend}
            className="text-white hover:underline mt-2"
            disabled={!!resendStatus}
          >
            {resendStatus || 'Resend Code'}
          </button>
        </div>
      </motion.div>
    </div>
  );
};

export default VerifyOtpPage;

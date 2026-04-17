import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { sendOtp } from '../features/user/services/userService';

const SendOtpPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleSendOtp = async () => {
    setLoading(true);
    setError(null);
    try {
      await sendOtp();
      navigate('/verify-otp');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send verification code. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-flixor-dark pt-20 px-4 relative">
      <div className="max-w-md w-full bg-black/80 border border-white/20 p-10 rounded-md shadow-2xl z-10 relative text-center">
        <h2 className="text-3xl font-bold mb-6 text-white">Verify Your Email</h2>
        
        <p className="text-flixor-lightGray mb-8">
          To finalize your account setup and secure your profile, please verify your email address. We'll send a 6-digit confirmation code.
        </p>

        {error && <div className="bg-flixor-red p-3 rounded mb-6 text-white text-sm">{error}</div>}

        <button 
          onClick={handleSendOtp}
          disabled={loading}
          className="w-full bg-flixor-red hover:bg-flixor-hover text-white font-bold py-4 rounded transition-colors text-lg shadow-lg disabled:opacity-50"
        >
          {loading ? 'Sending Code...' : 'Send Verification Code'}
        </button>
      </div>

      {/* Decorative gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-t from-flixor-dark to-transparent opacity-80 pointer-events-none"></div>
    </div>
  );
};

export default SendOtpPage;

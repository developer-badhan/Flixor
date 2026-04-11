import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '../hooks/useAuth';
import { useNavigate, Link } from 'react-router-dom';
import { register as registerService } from '../features/auth/services/authService';

const SignupPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const data = await registerService({ name, email, password });
      login(data.token);
      navigate('/send-otp');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to sign up. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div 
      className="min-h-screen flex items-center justify-center bg-cover bg-center relative"
      style={{ backgroundImage: 'url(https://assets.nflxext.com/ffe/siteui/vlv3/28b69a57-cadf-43d9-8a95-e5f2e11199de/3160e1fd-1c9f-4db3-96b4-2fc57ed7cb4c/US-en-20241008-TRIFECTA-perspective_2bf7aab7-d078-43ec-b6a8-a5a40bf5675c_small.jpg)' }}
    >
      <div className="absolute inset-0 bg-black/60 sm:bg-black/40"></div>
      
      <motion.div 
        className="w-full max-w-md mx-4 relative z-10"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
      >
        <div className="bg-black/75 p-10 sm:p-16 rounded-md shadow-2xl">
          <h2 className="text-3xl font-bold mb-8 text-white">Sign Up</h2>
          
          {error && <div className="bg-flixor-red p-3 rounded mb-4 text-white text-sm">{error}</div>}
          
          <form onSubmit={handleSignup} className="space-y-4">
             <div>
              <input 
                type="text" 
                placeholder="Full Name" 
                className="w-full bg-[#333333] text-white px-4 py-3 rounded focus:outline-none focus:ring-2 focus:ring-white/30"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div>
              <input 
                type="email" 
                placeholder="Email address" 
                className="w-full bg-[#333333] text-white px-4 py-3 rounded focus:outline-none focus:ring-2 focus:ring-white/30"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div>
              <input 
                type="password" 
                placeholder="Password" 
                className="w-full bg-[#333333] text-white px-4 py-3 rounded focus:outline-none focus:ring-2 focus:ring-white/30"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <button 
              type="submit" 
              className="w-full bg-flixor-red hover:bg-flixor-hover text-white font-bold py-3 rounded transition-colors mt-6"
              disabled={loading}
            >
              {loading ? 'Signing Up...' : 'Sign Up'}
            </button>
          </form>
          
          <div className="mt-16 text-[#8c8c8c]">
            <p>Already have an account? <Link to="/login" className="text-white hover:underline">Sign in.</Link></p>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default SignupPage;

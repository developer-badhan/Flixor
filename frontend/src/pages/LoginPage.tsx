import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useAuth } from '../hooks/useAuth';
import { useNavigate, Link } from 'react-router-dom';
import { login as loginService } from '../features/auth/services/authService';

const LoginPage: React.FC = () => {
  const [email, setEmail]       = useState('');
  const [password, setPassword] = useState('');
  const [error, setError]       = useState<string | null>(null);
  const [loading, setLoading]   = useState(false);

  /**
  * BUG FIX 1: was `const { login } = useAuth()`
  * The old useAuth() hook exposed a login(accessToken) method that only stored
  * the access token and never touched shared React state.
  * AuthContext exposes setTokens(accessToken, refreshToken) which:
  *   - updates the shared isAuthenticated state → ProtectedRoute reacts immediately
  *   - saves both tokens to localStorage
  */
  const { setTokens } = useAuth();

  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const tokens = await loginService({ email, password });
      /**
      * tokens is already the unwrapped TokenResponse:
      * { access_token, refresh_token, token_type, expires_in }
      * (api.ts interceptor unwraps response.data.data → response.data)
      * 
      * BUG FIX 2: was `login(tokens.accessToken)`
      * Three problems in one line:
      *   a) `login` no longer exists on context — replaced by `setTokens`
      *   b) `tokens.accessToken` is camelCase → undefined (backend sends access_token)
      *   c) Only the access token was passed — refresh token was never saved,
      *      so every token refresh call read null from localStorage and sent
      *      { refresh_token: null } to the backend → 400 → logout redirect
      */
      setTokens(tokens.access_token, tokens.refresh_token);   

      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to login. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center bg-cover bg-center relative"
      style={{
        backgroundImage:
          'url(https://assets.nflxext.com/ffe/siteui/vlv3/28b69a57-cadf-43d9-8a95-e5f2e11199de/3160e1fd-1c9f-4db3-96b4-2fc57ed7cb4c/US-en-20241008-TRIFECTA-perspective_2bf7aab7-d078-43ec-b6a8-a5a40bf5675c_small.jpg)',
      }}
    >
      <div className="absolute inset-0 bg-black/60 sm:bg-black/40" />

      <motion.div
        className="w-full max-w-md mx-4 relative z-10"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
      >
        <div className="bg-black/75 p-10 sm:p-16 rounded-md shadow-2xl">
          <h2 className="text-3xl font-bold mb-8 text-white">Sign In</h2>

          {error && (
            <div className="bg-flixor-red p-3 rounded mb-4 text-white text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <input
                type="email"
                placeholder="Email or phone number"
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
              {loading ? 'Signing In...' : 'Sign In'}
            </button>
            <div className="flex justify-between items-center text-[#b3b3b3] text-sm mt-2">
              <label className="flex items-center gap-2">
                <input type="checkbox" className="accent-gray-500" /> Remember me
              </label>
              <Link to="/help" className="hover:underline">
                Need help?
              </Link>
            </div>
          </form>

          <div className="mt-16 text-[#8c8c8c]">
            <p>
              New to Flixor?{' '}
              <Link to="/signup" className="text-white hover:underline">
                Sign up now.
              </Link>
            </p>
            <p className="text-xs mt-4">
              This page is protected by Google reCAPTCHA to ensure you're not a bot.{' '}
              <Link to="/learn-more" className="text-blue-500 hover:underline">
                Learn more.
              </Link>
            </p>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default LoginPage;
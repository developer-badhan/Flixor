import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

/** 
 * BUG FIX: AuthProvider was missing entirely.
 * Without it, useAuthContext() throws "must be used inside <AuthProvider>"
 * on every component that calls useAuth() — including ProtectedRoute.
 * Wrapping here means the single shared auth state covers the whole app.
*/

import { AuthProvider } from './context/AuthContext'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <App />
    </AuthProvider>
  </StrictMode>,
)
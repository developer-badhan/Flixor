import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Search, Bell, User, BarChart2, Sparkles, LogOut, Settings } from 'lucide-react';
import { useUserContext } from '../context/UserContext';
import { useAuthContext } from '../context/AuthContext';
import logo from '../assets/logo.png';

// ─── Nav link definitions ─────────────────────────────────────────────────────

const NAV_LINKS = [
  { to: '/',               label: 'Home' },
  { to: '/movies',         label: 'Movies' },
  { to: '/watchlist',      label: 'My List' },
  { to: '/recommendations', label: 'For You',    icon: <Sparkles size={13} /> },
  { to: '/analytics',      label: 'Analytics',   icon: <BarChart2 size={13} /> },
];

// ─── Component ────────────────────────────────────────────────────────────────

const Navbar: React.FC = () => {
  const [isScrolled, setIsScrolled]         = useState(false);
  const [isSearchActive, setIsSearchActive] = useState(false);
  const [searchQuery, setSearchQuery]       = useState('');
  const navigate  = useNavigate();
  const location  = useLocation();
  const { user } = useUserContext();
  const { clearTokens } = useAuthContext();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 0);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/search?title=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  return (
    <nav
      className={`fixed w-full z-50 transition-all duration-300 ${
        isScrolled
          ? 'bg-flixor-dark shadow-md'
          : 'bg-transparent bg-gradient-to-b from-black/80 to-transparent'
      }`}
    >
      <div className="max-w-[1920px] mx-auto px-4 md:px-8 py-4 flex items-center justify-between">

        {/* Left: Logo + Nav links */}
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <img src={logo} alt="Flixor Logo" className="h-8 md:h-10" />
          </Link>

          <div className="hidden md:flex items-center gap-0.5">
            {NAV_LINKS.map(({ to, label, icon }) => {
              const isActive = location.pathname === to;
              return (
                <Link
                  key={to}
                  to={to}
                  className={`
                    relative flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium
                    transition-all duration-200
                    ${isActive
                      ? 'text-white bg-white/8'
                      : 'text-flixor-lightGray hover:text-white hover:bg-white/5'
                    }
                  `}
                >
                  {icon && (
                    <span
                      style={{
                        color: isActive
                          ? to === '/recommendations' ? '#a855f7'
                            : to === '/analytics' ? '#e50914'
                            : 'inherit'
                          : 'inherit',
                      }}
                    >
                      {icon}
                    </span>
                  )}
                  {label}
                  {isActive && (
                    <span
                      className="absolute -bottom-px left-3 right-3 h-px rounded-full"
                      style={{
                        background:
                          to === '/recommendations'
                            ? '#a855f7'
                            : to === '/analytics'
                            ? '#e50914'
                            : '#fff',
                      }}
                    />
                  )}
                </Link>
              );
            })}
          </div>
        </div>

        {/* Right: Search + Bell + Profile */}
        <div className="flex items-center gap-4">
          <form onSubmit={handleSearchSubmit} className="relative flex items-center justify-end">
            <div
              className={`flex items-center bg-black/80 border transition-all duration-300 ${
                isSearchActive
                  ? 'border-white/80 w-64 px-2'
                  : 'border-transparent w-8 px-0'
              }`}
            >
              <button
                type="button"
                onClick={() =>
                  isSearchActive && !searchQuery
                    ? setIsSearchActive(false)
                    : setIsSearchActive(true)
                }
                className="text-flixor-lightGray hover:text-white transition-colors flex-shrink-0"
              >
                <Search size={22} />
              </button>
              <input
                type="text"
                placeholder="Title, genre"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onBlur={() => !searchQuery && setIsSearchActive(false)}
                className={`bg-transparent text-sm text-white px-2 py-1.5 outline-none w-full transition-all duration-300 ${
                  isSearchActive ? 'opacity-100' : 'opacity-0 pointer-events-none'
                }`}
                autoFocus={isSearchActive}
              />
            </div>
          </form>

          <button className="hover:text-flixor-lightGray transition-colors hidden md:block">
            <Bell size={20} />
          </button>

          <div
            className="relative"
            onMouseEnter={() => setIsDropdownOpen(true)}
            onMouseLeave={() => setIsDropdownOpen(false)}
          >
            <div className="w-8 h-8 rounded-full overflow-hidden border border-transparent hover:border-white transition-all cursor-pointer bg-flixor-dark flex items-center justify-center">
              <img
                src={user?.profile_picture || "https://cdn.iconscout.com/icon/premium/png-256-thumb/user-icon-svg-download-png-5358.png?f=webp&w=128"}
                alt="Profile"
                className="w-full h-full object-cover"
                onError={(e) => {
                  (e.target as HTMLImageElement).src = "https://cdn.iconscout.com/icon/premium/png-256-thumb/user-icon-svg-download-png-5358.png?f=webp&w=128";
                }}
              />
            </div>

            {/* Dropdown Menu */}
            <div
              className={`absolute right-0 mt-2 w-48 bg-flixor-dark/95 backdrop-blur-md border border-white/10 rounded-xl shadow-2xl py-2 transition-all duration-300 origin-top-right ${
                isDropdownOpen
                  ? 'opacity-100 scale-100 translate-y-0 visible'
                  : 'opacity-0 scale-95 -translate-y-2 invisible'
              }`}
            >
              <div className="px-4 py-2 border-b border-white/10 mb-2">
                <p className="text-sm text-white font-medium truncate">{user?.username || 'Guest'}</p>
                <p className="text-xs text-gray-400 truncate mt-0.5">{user?.email || 'Not logged in'}</p>
              </div>

              <Link
                to="/me"
                className="flex items-center gap-3 px-4 py-2 text-sm text-gray-300 hover:text-white hover:bg-white/10 transition-colors"
                onClick={() => setIsDropdownOpen(false)}
              >
                <User size={16} />
                Profile
              </Link>

              <button
                onClick={() => {
                  setIsDropdownOpen(false);
                  clearTokens();
                  navigate('/login');
                }}
                className="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-300 hover:text-white hover:bg-white/10 transition-colors"
              >
                <LogOut size={16} />
                Sign out
              </button>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
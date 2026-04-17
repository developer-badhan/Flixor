import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Search, Bell, User, BarChart2 } from 'lucide-react';
import logo from '../assets/logo.png';


const NAV_LINKS = [
  { to: '/',          label: 'Home' },
  { to: '/movies',    label: 'Movies' },
  { to: '/watchlist', label: 'My List' },
  { to: '/analytics', label: 'Analytics', icon: <BarChart2 size={14} /> },
];

const Navbar: React.FC = () => {
  const [isScrolled, setIsScrolled]       = useState(false);
  const [isSearchActive, setIsSearchActive] = useState(false);
  const [searchQuery, setSearchQuery]     = useState('');
  const navigate  = useNavigate();
  const location  = useLocation();

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
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <img src={logo} alt="Flixor Logo" className="h-8 md:h-10" />
          </Link>
 
          <div className="hidden md:flex items-center gap-1">
            {NAV_LINKS.map(({ to, label, icon }) => {
              const isActive = location.pathname === to;
              return (
                <Link
                  key={to}
                  to={to}
                  className={`
                    flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200
                    ${isActive
                      ? 'text-white bg-white/10'
                      : 'text-flixor-lightGray hover:text-white hover:bg-white/5'
                    }
                  `}
                >
                  {icon && <span className={isActive ? 'text-flixor-red' : ''}>{icon}</span>}
                  {label}
                  {/* Active underline dot */}
                  {isActive && (
                    <span className="ml-0.5 w-1 h-1 rounded-full bg-flixor-red inline-block" />
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
 
          <Link
            to="/profile"
            className="flex items-center gap-2 hover:text-flixor-lightGray transition-colors cursor-pointer"
          >
            <div className="w-8 h-8 bg-flixor-gray rounded flex items-center justify-center">
              <User size={20} />
            </div>
          </Link>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;

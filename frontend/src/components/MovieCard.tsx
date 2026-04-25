import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Play } from 'lucide-react';

/** 
 * BUG FIX: onError previously called `style.display = 'none'` on the <img>.
 * This hid the element but the fallback div was in the else branch —
 * once the img rendered (even briefly), the else branch was gone.
 * Result: broken image → blank gray box with no text visible.
 *
 * FIX: track image error with useState → when error fires, React re-renders
 * and switches to the title fallback div. Same pattern as Thumbnail in MostWatched.
*/

/**
 * The movie card component for the application.
 * @param id - The ID of the movie.
 * @param title - The title of the movie.
 * @param posterUrl - The URL of the movie poster.
 * @returns A React component that represents the movie card of the application.
 * @component
 * @returns {React.FC<MovieCardProps>}
 */
interface MovieCardProps {
  id: string;
  title: string;
  posterUrl: string;
}

const MovieCard: React.FC<MovieCardProps> = ({ id, title, posterUrl }) => {
  const [imgError, setImgError] = useState(false);

  // Show fallback if: no URL provided, OR image failed to load
  const showFallback = !posterUrl || imgError;

  return (
    <Link to={`/movie/${id}`}>
      <motion.div
        className="relative rounded-md overflow-hidden aspect-[2/3] group cursor-pointer bg-flixor-gray"
        whileHover={{ scale: 1.05 }}
        transition={{ duration: 0.2 }}
      >
        {showFallback ? (
          // Visible fallback — never an empty box
          <div className="w-full h-full flex items-center justify-center bg-black/40 p-4 text-center">
            <span className="text-sm font-bold text-white drop-shadow-md leading-tight line-clamp-4">
              {title}
            </span>
          </div>
        ) : (
          <img
            src={posterUrl}
            alt={title}
            className="w-full h-full object-cover brightness-90 group-hover:brightness-50 transition-all duration-300"
            onError={() => setImgError(true)}   // FIX: flips to title fallback on broken URL
          />
        )}

        {/* Hover overlay (always shown regardless of image state) */}
        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center">
          <motion.div
            initial={{ y: 10, opacity: 0 }}
            whileHover={{ y: 0, opacity: 1 }}
            className="flex flex-col items-center gap-2"
          >
            <div className="bg-flixor-red p-3 rounded-full">
              <Play size={24} fill="white" />
            </div>
            <p className="text-sm font-bold text-center px-2 text-white line-clamp-2">{title}</p>
          </motion.div>
        </div>
      </motion.div>
    </Link>
  );
};

export default MovieCard;
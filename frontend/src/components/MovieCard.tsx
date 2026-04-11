import React from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Play } from 'lucide-react';

interface MovieCardProps {
  id: string;
  title: string;
  posterUrl: string;
}

const MovieCard: React.FC<MovieCardProps> = ({ id, title, posterUrl }) => {
  return (
    <Link to={`/movie/${id}`}>
      <motion.div
        className="relative rounded-md overflow-hidden aspect-[2/3] group cursor-pointer bg-flixor-gray"
        whileHover={{ scale: 1.05 }}
        transition={{ duration: 0.2 }}
      >
        {posterUrl ? (
          <img 
            src={posterUrl} 
            alt={title} 
            className="w-full h-full object-cover brightness-90 group-hover:brightness-50 transition-all duration-300"
            onError={(e) => {
              // Fallback to text if image fails to load
              (e.target as HTMLImageElement).style.display = 'none';
            }}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-black/40 p-4 text-center">
            <span className="text-sm font-bold text-white drop-shadow-md">{title}</span>
          </div>
        )}
        
        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center">
          <motion.div
            initial={{ y: 10, opacity: 0 }}
            whileHover={{ y: 0, opacity: 1 }}
            className="flex flex-col items-center"
          >
            <div className="bg-flixor-red p-3 rounded-full mb-2">
               <Play size={24} fill="white" />
            </div>
            <p className="text-sm font-bold text-center px-2">{title}</p>
          </motion.div>
        </div>
      </motion.div>
    </Link>
  );
};

export default MovieCard;

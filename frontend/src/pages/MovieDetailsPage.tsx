import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Play, Plus, ThumbsUp, Heart } from 'lucide-react';
import { useFetch } from '../hooks/useFetch';

const MovieDetailsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data: movie, loading, error } = useFetch<any>(`/movies/${id}`);
  const [isPlaying, setIsPlaying] = useState(false);

  if (loading) {
    return <div className="h-screen bg-flixor-dark flex items-center justify-center">Loading...</div>;
  }

  if (error || !movie) {
    return <div className="h-screen bg-flixor-dark flex items-center justify-center text-red-500">Error loading movie</div>;
  }

  return (
    <div className="min-h-screen bg-flixor-dark">
      {isPlaying ? (
        <div className="w-full h-screen bg-black flex items-center justify-center pt-20">
          <video 
             className="w-full h-full object-contain" 
             controls 
             autoPlay 
             src={movie.stream_url || ''} // Handle empty per requirements
             aria-label="Video Player"
          >
            Your browser does not support the video tag.
          </video>
          {/* Overlay to exit could be added here */}
          <button 
             className="absolute top-24 right-8 bg-white/20 p-2 rounded-full hover:bg-white/40 z-50 text-white"
             onClick={() => setIsPlaying(false)}
          >
            Close
          </button>
        </div>
      ) : (
        <div className="relative pb-20">
          {/* Backdrop */}
          <div className="relative h-[60vh] w-full">
            <img 
              src={'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?q=80&w=2070'} 
              alt={movie.title} 
              className="w-full h-full object-cover opacity-30"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-flixor-dark to-transparent"></div>
          </div>

          {/* Details Content */}
          <div className="max-w-6xl mx-auto px-8 -mt-32 relative z-10">
            <motion.div 
               className="flex flex-col md:flex-row gap-8"
               initial={{ opacity: 0, y: 50 }}
               animate={{ opacity: 1, y: 0 }}
               transition={{ duration: 0.5 }}
            >
              <img 
                src={movie.thumbnail_url || 'https://via.placeholder.com/300x450'} 
                alt={movie.title} 
                className="w-64 rounded-lg shadow-2xl border border-flixor-gray hidden md:block"
              />
              <div className="flex-1">
                <h1 className="text-4xl md:text-5xl font-bold mb-4">{movie.title}</h1>
                <div className="flex items-center gap-4 text-flixor-lightGray mb-6 text-sm font-medium">
                  <span>{movie.year || '2023'}</span>
                  <span className="border border-flixor-gray px-1 rounded">{movie.age_rating || '16+'}</span>
                  <span>{movie.duration ? `${movie.duration} min` : '120 min'}</span>
                  <span className="text-flixor-red font-bold">HD</span>
                </div>
                
                <p className="text-lg mb-8 leading-relaxed max-w-3xl">
                  {movie.description}
                </p>

                <div className="flex items-center gap-4">
                  <button 
                     className="flex items-center gap-2 bg-white text-black px-8 py-3 rounded drop-shadow hover:bg-opacity-80 transition-all font-bold text-lg"
                     onClick={() => setIsPlaying(true)}
                  >
                    <Play fill="currentColor" /> Play
                  </button>
                  <button className="p-3 border-2 border-gray-500 rounded-full hover:border-white transition-colors bg-[#2a2a2a]/60 backdrop-blur" title={"Add to watchlist"}>
                    <Plus />
                  </button>
                  <button className="p-3 border-2 border-gray-500 rounded-full hover:border-white transition-colors bg-[#2a2a2a]/60 backdrop-blur">
                    <ThumbsUp />
                  </button>
                  <button className="p-3 border-2 border-gray-500 rounded-full hover:border-white transition-colors bg-[#2a2a2a]/60 backdrop-blur">
                    <Heart />
                  </button>
                </div>

                <div className="mt-8 space-y-2 text-sm">
                  <p><span className="text-flixor-lightGray">Genres:</span> {movie.genres?.join(', ') || 'Action, Thriller'}</p>
                  <p><span className="text-flixor-lightGray">Director:</span> {movie.director || 'Unknown'}</p>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MovieDetailsPage;

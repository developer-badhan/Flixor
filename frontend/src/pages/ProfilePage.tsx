import React from 'react';
import { useFetch } from '../hooks/useFetch';

const ProfilePage: React.FC = () => {
  const { data: user, loading, error } = useFetch<any>('/user/me');

  return (
    <div className="pt-24 px-8 min-h-screen bg-flixor-dark">
      <h2 className="text-3xl font-bold mb-8 text-white">Profile</h2>
      <div className="glass p-8 rounded-lg max-w-2xl mx-auto border border-flixor-gray">
        {loading ? (
          <div className="animate-pulse space-y-4">
            <div className="h-6 bg-flixor-gray rounded w-1/3"></div>
            <div className="h-6 bg-flixor-gray rounded w-1/2"></div>
            <div className="h-6 bg-flixor-gray rounded w-1/4"></div>
          </div>
        ) : error ? (
          <p className="text-flixor-red">Failed to load profile. Please login.</p>
        ) : user ? (
          <div className="space-y-6 text-lg text-flixor-lightGray">
            <div className="flex items-center gap-4 mb-8">
              <div className="w-20 h-20 bg-flixor-gray rounded-md flex items-center justify-center text-3xl font-bold text-white">
                {user.name ? user.name.charAt(0).toUpperCase() : 'U'}
              </div>
              <div>
                <h3 className="text-2xl font-bold text-white mb-1">{user.name}</h3>
                <p>{user.email}</p>
              </div>
            </div>
            
            <div className="bg-black/50 p-6 rounded border border-flixor-gray">
              <p className="mb-2"><strong className="text-white">Account ID:</strong> {user.id}</p>
              <p className="mb-2"><strong className="text-white">Joined:</strong> {new Date(user.created_at).toLocaleDateString()}</p>
              <p><strong className="text-white">Status:</strong> Active</p>
            </div>
            
            <div className="mt-8">
              <button className="bg-white text-black font-bold py-2 px-6 rounded hover:bg-gray-200 transition-colors">
                Edit Profile
              </button>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
};

export default ProfilePage;

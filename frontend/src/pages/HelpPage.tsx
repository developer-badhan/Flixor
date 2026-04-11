import React from 'react';
import { motion } from 'framer-motion';

const HelpPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-flixor-dark pt-24 px-8 md:px-16 pb-20">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-4xl mx-auto"
      >
        <h1 className="text-4xl font-bold mb-8 text-white">Help Center</h1>
        
        <div className="space-y-6">
          <section className="bg-[#141414] border border-flixor-gray p-6 rounded-md shadow-lg">
            <h2 className="text-xl font-bold text-white mb-3">How do I verify my email?</h2>
            <p className="text-flixor-lightGray">
              Once you sign up, you will be redirected to the Verification page. Click "Send Verification Code" and check your inbox for a 6-digit number. Enter it on the next screen to fully unlock your account!
            </p>
          </section>

          <section className="bg-[#141414] border border-flixor-gray p-6 rounded-md shadow-lg">
            <h2 className="text-xl font-bold text-white mb-3">I forgot my password</h2>
            <p className="text-flixor-lightGray">
              Currently, password resets are handled via our system administrators. Please contact support@flixor.com with your registered email address to request a reset link.
            </p>
          </section>

          <section className="bg-[#141414] border border-flixor-gray p-6 rounded-md shadow-lg">
            <h2 className="text-xl font-bold text-white mb-3">Is Flixor free?</h2>
            <p className="text-flixor-lightGray">
              Yes, Flixor utilizes public domain repositories like the Internet Archive to bring you classic and indie movies completely free of charge. Your view history and watchlists are saved securely.
            </p>
          </section>
        </div>

        <div className="mt-12 text-center">
          <p className="text-flixor-lightGray">Still need help?</p>
          <a href="mailto:support@flixor.com" className="text-flixor-red font-bold hover:underline">Contact Support</a>
        </div>
      </motion.div>
    </div>
  );
};

export default HelpPage;

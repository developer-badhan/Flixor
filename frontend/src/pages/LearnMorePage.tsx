import React from 'react';
import { motion } from 'framer-motion';

const LearnMorePage: React.FC = () => {
  return (
    <div className="min-h-screen bg-flixor-dark pt-24 px-8 md:px-16 pb-20">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-4xl mx-auto"
      >
        <h1 className="text-4xl font-bold mb-8 text-white">Bot Protection & Privacy</h1>
        
        <div className="prose prose-invert max-w-none text-flixor-lightGray">
          <h3 className="text-2xl text-white font-semibold mb-4 mt-8">Google reCAPTCHA</h3>
          <p className="mb-6 leading-relaxed">
            This page is protected by Google reCAPTCHA to ensure you're not a bot. This information is used solely to verify human interaction securely and prevent abuse of our unified cinematic services. 
          </p>

          <h3 className="text-2xl text-white font-semibold mb-4 mt-8">How it Works</h3>
          <p className="mb-6 leading-relaxed">
            The reCAPTCHA API works by collecting hardware and software information, such as device and application data, and sending these data to Google for analysis. The information collected in connection with your use of the service will be used for improving reCAPTCHA and for general security purposes.
          </p>

          <h3 className="text-2xl text-white font-semibold mb-4 mt-8">Data Privacy</h3>
          <p className="mb-6 leading-relaxed">
            It will not be used for personalized advertising by Google. We respect your security and ensuring that our systems remain fast, durable, and reliable requires these protective measures. By interacting with the site, you acknowledge Google's strict Privacy Policy and Terms of Service regarding bot protection.
          </p>

          <div className="mt-12 p-6 border border-flixor-gray bg-[#141414] rounded-md">
            <p className="text-sm font-medium">
              Read Google's <a href="https://policies.google.com/privacy" target="_blank" rel="noreferrer" className="text-blue-500 hover:underline">Privacy Policy</a> and <a href="https://policies.google.com/terms" target="_blank" rel="noreferrer" className="text-blue-500 hover:underline">Terms of Service</a> for detailed legal documentation.
            </p>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default LearnMorePage;

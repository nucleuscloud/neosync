const { withContentlayer } = require('next-contentlayer');

require('./env.js');

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'assets.nucleuscloud.com',
        port: '',
        pathname: '/**',
      },
    ],
  },
};
module.exports = withContentlayer(nextConfig);

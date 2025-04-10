import path from 'path';
import * as url from 'url';
const __dirname = url.fileURLToPath(new URL('.', import.meta.url));

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  output: 'standalone',
  outputFileTracingRoot: path.join(__dirname, '../../'),
  experimental: process.env.NODEJS_PROXY_TIMEOUT
    ? {
        proxyTimeout: parseInt(process.env.NODEJS_PROXY_TIMEOUT),
      }
    : undefined,
};

export default config;

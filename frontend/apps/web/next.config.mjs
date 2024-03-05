import path from 'path';
import * as url from 'url';
const __dirname = url.fileURLToPath(new URL('.', import.meta.url));

export default {
  reactStrictMode: true,
  output: 'standalone',
  transpilePackages: [
    '@dopt/react-checklist',
    '@dopt/react-theme',
    '@dopt/react-rich-text',
    '@dopt/core-theme',
    '@dopt/core-rich-text',
  ],
  experimental: {
    outputFileTracingRoot: path.join(__dirname, '../../'),
  },
};

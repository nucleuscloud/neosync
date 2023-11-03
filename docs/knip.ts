import type { KnipConfig } from 'knip';

const config: KnipConfig = {
  ignore: [
    'babel.config.js',
    'docusaurus.config.js',
    'sidebars.js',
    'static/sync-dark-mode.js',
    'src/components/HomepageFeatures/**',
    'src/theme/**',
  ],
  ignoreDependencies: [
    '@docusaurus/preset-classic',
    '@mdx-js/react',
    '@radix-ui/react-icons',
    'autoprefixer',
    'postcss',
    'prism-react-renderer',
    'react-icons',
    'tailwindcss',
  ],
  ignoreBinaries: [],
};

export default config;

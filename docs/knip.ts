const config = {
  ignore: [
    'babel.config.js',
    'docusaurus.config.js',
    'sidebars.js',
    'static/sync-dark-mode.js',
    'src/CustomComponents/**',
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

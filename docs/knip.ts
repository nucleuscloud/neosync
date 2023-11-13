const config = {
  ignore: [
    'babel.config.js',
    'docusaurus.config.js',
    'sidebars.js',
    'static/sync-dark-mode.js',
    'src/**',
    'src/theme/**',
  ],
  ignoreDependencies: [
    '@docusaurus/preset-classic',
    '@mdx-js/react',
    '@radix-ui/react-icons',
    'autoprefixer',
    'postcss',
    'node',
    '@docusaurus/theme-classic',
    'prism-react-renderer',
    'react-icons',
    'tailwindcss',
    'class-variance-authority',
    'tailwind-merge',
  ],
  ignoreBinaries: [],
};

export default config;

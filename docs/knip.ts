const config = {
  ignore: [
    'babel.config.js',
    'docusaurus.config.ts',
    'sidebars.ts',
    'proto-sidebars.ts',
    'protos/**',
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
    'prism-react-renderer',
    'react-icons',
    'tailwindcss',
    'class-variance-authority',
    'tailwind-merge',
    'docusaurus-protobuffet',
  ],
  ignoreBinaries: [],
};

export default config;

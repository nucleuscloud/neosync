import type { KnipConfig } from 'knip';

const config: KnipConfig = {
  ignore: [
    'next-env.d.ts',
    'next.config.mjs',
    'neosync-api-client',
    'components/ui',
    'postcss.config.cjs',
  ],
  ignoreDependencies: [
    '@radix-ui/react-accordion',
    '@radix-ui/react-avatar',
    '@radix-ui/react-checkbox',
    '@radix-ui/react-dialog',
    '@radix-ui/react-label',
    '@radix-ui/react-popover',
    '@radix-ui/react-scroll-area',
    '@radix-ui/react-select',
    '@radix-ui/react-separator',
    '@radix-ui/react-slot',
    '@radix-ui/react-switch',
    '@radix-ui/react-toast',
    'class-variance-authority',
    'cmdk',
    'autoprefixer',
    'postcss',
    'eslint-config-next',
  ],
  ignoreBinaries: ['tail'],
};

export default config;

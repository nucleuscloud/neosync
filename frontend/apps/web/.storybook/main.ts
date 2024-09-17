import type { StorybookConfig } from '@storybook/nextjs';
import TsconfigPathsPlugin from 'tsconfig-paths-webpack-plugin';

import { dirname, join, resolve } from 'path';

/**
 * This function is used to resolve the absolute path of a package.
 * It is needed in projects that use Yarn PnP or are set up within a monorepo.
 */
function getAbsolutePath(value: string): any {
  return dirname(require.resolve(join(value, 'package.json')));
}
const config: StorybookConfig = {
  stories: [
    // '../stories/**/*.mdx',
    // '../stories/**/*.stories.@(js|jsx|mjs|ts|tsx)',
    '../components/**/*.stories.@(js|jsx|mjs|ts|tsx)',
  ],
  addons: [
    getAbsolutePath('@storybook/addon-onboarding'),
    getAbsolutePath('@storybook/addon-links'),
    getAbsolutePath('@storybook/addon-essentials'),
    getAbsolutePath('@chromatic-com/storybook'),
    getAbsolutePath('@storybook/addon-interactions'),
    getAbsolutePath('@storybook/addon-themes'),
    getAbsolutePath('@storybook/addon-webpack5-compiler-swc'),
  ],
  framework: {
    name: getAbsolutePath('@storybook/nextjs'),
    options: {},
  },
  // webpackFinal: (config) => {
  //   if (config.resolve && config.resolve.alias) {
  //     config.resolve.alias['@neosync/sdk'] = path.resolve(
  //       __dirname,
  //       '../../packages/sdk'
  //     );
  //     // config.resolve.alias['@'] = path.resolve(__dirname, '../');
  //   }
  //   return config;
  // },
  webpackFinal: async (config) => {
    if (config.resolve) {
      config.resolve.plugins = [
        ...(config.resolve.plugins || []),
        new TsconfigPathsPlugin({
          extensions: config.resolve.extensions,
          configFile: resolve(__dirname, '../tsconfig.json'),
        }),
      ];
      // config.resolve.alias = {
      //   ...config.resolve.alias,
      //   '@neosync/sdk': resolve(__dirname, '../../../packages/sdk'),
      // };
      config.resolve.alias = {
        ...config.resolve.alias,
        '@neosync/sdk': resolve(__dirname, '../../sdk/ts-client'),
      };
    }
    return config;
  },
};
export default config;

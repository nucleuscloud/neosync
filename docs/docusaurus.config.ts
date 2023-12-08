// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import type { Config } from '@docusaurus/types';
import { themes } from 'prism-react-renderer';

const config: Config = {
  title: 'Neosync',
  tagline: 'Open source Test Data Management',
  favicon: 'img/logo_light_mode.png',

  // Set the production url of your s here
  url: 'https://docs.neosync.dev',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'nucleuscloud', // Usually your GitHub org/user name.
  projectName: 'neosync', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn', //should probably be throw or warn but was causing a known issue in the markdown parsing of readme files from node_modules. https://github.com/facebook/docusaurus/issues/6370

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  plugins: [
    async function tailwindcssPlugin(context, options) {
      return {
        name: 'docusaurus-tailwindcss',
        configurePostCss(postcssOptions) {
          // Appends TailwindCSS and AutoPrefixer.
          postcssOptions.plugins.push(require('tailwindcss'));
          postcssOptions.plugins.push(require('autoprefixer'));
          return postcssOptions;
        },
      };
    },
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/nucleuscloud/neosync/blob/main/docs',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
    [
      'docusaurus-protobuffet',
      {
        protobuffet: {
          fileDescriptorsPath: './protos/data/proto_docs.json',
          protoDocsPath: 'protos',
          sidebarPath: './protos/proto-sidebars.js',
        },
        docs: {
          routeBasePath: 'api',
          sidebarPath: './proto-sidebars.ts',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],

  themeConfig: {
    image: 'img/docusaurus-social-card.jpg',
    colorMode: {
      defaultMode: 'light',
      // disabling color mode and preference until dark mode switching is fixed
      disableSwitch: true,
      respectPrefersColorScheme: false,
    },
    navbar: {
      logo: {
        alt: 'Neosync',
        srcDark: 'img/logo_and_text_dark_mode.png',
        src: 'img/logo_and_text_light_mode.png',
      },

      items: [
        {
          type: 'custom-Gitlink',
          position: 'right',
        },
        { to: '/', label: 'Docs' },
        { to: '/protos', label: 'API' },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `Copyright © Nucleus Cloud Corp ${new Date().getFullYear()}`,
    },
    prism: {
      theme: themes.github,
      darkTheme: themes.dracula,
    },
  },
};

export default config;

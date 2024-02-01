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
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/nucleuscloud/neosync/blob/main/docs',
        },
        blog: {
          routeBasePath: '/changelog',
          editUrl: 'https://github.com/nucleuscloud/neosync/blob/main/docs',
          blogTitle: 'Neosync Changelog',
          blogDescription: 'Neosync Changelog',
          blogSidebarTitle: ' Changelog',
          blogSidebarCount: 'ALL',
        },
        theme: {
          customCss: ['./src/css/custom.css'],
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
    image: 'img/docsOGimage.png',
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      // disabling preference until dark mode switching is fixed: https://github.com/facebook/docusaurus/issues/8938
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
          href: 'https://github.com/nucleuscloud/neosync',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
        {
          href: 'https://discord.com/invite/MFAMgnp4HF',
          position: 'right',
          className: 'header-discord-link',
          'aria-label': 'Discord Server',
        },
        { to: '/', label: 'Docs' },
        { to: '/api', label: 'API' },
        { to: '/changelog', label: 'Changelog' },
      ],
    },
    footer: {
      copyright: `Copyright © Nucleus Cloud Corp ${new Date().getFullYear()}`,
    },
    prism: {
      theme: themes.github,
      darkTheme: themes.dracula,
    },
    algolia: {
      appId: 'LUROM0SS2F',
      apiKey: 'a58584698f0541be72a223f5b33d59a9',
      indexName: 'neosync',
      contextualSearch: true,
      searchParameters: {},
      searchPagePath: 'search',
    },
  },
};

export default config;

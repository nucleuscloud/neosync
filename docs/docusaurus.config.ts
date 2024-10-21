// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import type { Config } from '@docusaurus/types';
import { themes } from 'prism-react-renderer';
import type * as Redocusaurus from 'redocusaurus';

const config: Config = {
  title: 'Neosync',
  tagline: 'Open source Data Anonymization and Synthetic Data',
  favicon: 'img/logo_light_mode.png',
  headTags: [
    {
      tagName: 'script',
      attributes: {
        id: 'koala-snippet',

        innerHTML: `!function(t){if(window.ko)return;window.ko=[],["identify","track","removeListeners","open","on","off","qualify","ready"].forEach(function(t){ko[t]=function(){var n=[].slice.call(arguments);return n.unshift(t),ko.push(n),ko}});var n=document.createElement("script");n.async=!0,n.setAttribute("src","https://cdn.getkoala.com/v1/pk_4fa92236b6fe5d23fb878c88c14d209fd48e/sdk.js"),(document.body || document.head).appendChild(n)}();`,
      },
    },
  ],
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
    [
      'posthog-docusaurus',
      {
        apiKey: process.env.POSTHOG_KEY
          ? process.env.POSTHOG_KEY
          : 'phc_2hFE16FGvpOmUdgVczrxrJPDJ1sp724se5w7uAte9GS',
        appUrl: process.env.POSTHOG_HOST
          ? process.env.POSTHOG_HOST
          : 'https://app.posthog.com',
        enableInDevelopment: false,
      },
    ],
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
          id: 'default',
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
          fileDescriptorsPath: './protos/proto_docs.json',
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
    [
      'redocusaurus',
      {
        specs: [
          // we just pass in the final generated file
          // from this package -> https://www.npmjs.com/package/openapi-merge-cli
          // tried this with both the redocly cli and the this open-api merge cli and the page won't render
          // once you add more than like 3 services, it just crashes
          // the files look fine so im not sure why, it looks like it's throwing some javascript heap memory error
          // then the javascript builds during npm run build
          // so it might be a plugin issue with redocusaurus but the example repos have more APIs than we do and they work fine
          // so im not sure why it's not working
          // for example, the swagger editor is able to render it just fine -> https://editor.swagger.io/?_gl=1*139d7v5*_gcl_au*MTQ1ODQ5MDk5My4xNzI5MTk5NTc0
          {
            spec: './merged_openapi.yaml',
            route: '/reference',
          },
        ],
        // Theme Options for modifying how redoc renders them
        theme: {
          // Change with your site colors
          primaryColor: '#1890ff',
        },
      },
    ] satisfies Redocusaurus.PresetEntry,
  ],

  themeConfig: {
    metadata: [
      {
        name: 'keywords',
        content:
          'open source, anonymization, data anonymization, synthetic data, data privacy, data security',
      },
    ],
    image: 'img/docsOG.png',
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
        { to: '/reference', label: 'Reference' },
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

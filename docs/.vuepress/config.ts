const { description } = require('../package')

import { defineUserConfig } from 'vuepress'

import {MixThemeConfig} from 'vuepress-theme-mix/lib/node'

export default defineUserConfig<MixThemeConfig>({
  title: null,
  description: description,
  head: [
    ['meta', { name: 'theme-color', content: '#803eaf' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }]
  ],
  base: "/osmosis/",

  locales: {
    '/': {
      lang: 'en-US',
      title: 'Osmosis Documentation', // for broswer tabs
      description: '',
    }
  },

  theme: 'vuepress-theme-mix',

  themeConfig: {
    logo: 'OSMOLogoTitleDark.png',
    logoDark: 'OSMOLogoTitleLight.png',
    title: 'Documentation', // for navbar
    docsRepo: 'osmosis-labs/osmosis',
    docsDir: 'docs',
    editLink: true,
    lastUpdated: true,

    locales: {
      '/': {
        navbar: [
          {
            text: 'Home',
            link: '/',
          },
          {
            text: 'Develop',
            link: '/test',
          },
          {
            text: 'Validate',
            link: '/validators',
          },
          {
            text: 'Integrate',
            link: '/integrate',
          },
          {
            text: 'Learn',
            link: '/introduction',
          },
          // {
          //   text: 'Reference',
          //   children: [
          //     '/configuration/',
          //     '/plugins.md',
          //     {
          //       text: 'Changelog',
          //       link: 'https://github.com/',
          //     },
          //   ],
          // },
          {
            text: 'Chat',
            link: 'https://v2.vuepress.vuejs.org/',
          },
          {
            text: 'GitHub',
            link: 'https://github.com/osmosis-labs/osmosis',
          },
        ],
        sidebar: {
          '/': [
            {
              type: 'group',
              text: 'About',
              link: '',
              children: [
                '/intro/',
                '/intro/wallets',
                '/intro/osmo',
                '/intro/terminology',
                '/intro/governance',
              ],
            },
            {
              type: 'group',
              text: 'Liquidity',
              link: '',
              children: [
                '/liquidity',
                '/liquidity/liquidity-bootstraping',
              ],
            },
            {
              type: 'group',
              text: 'Command Line',
              link: '',
              children: [
                '/cli',
                '/cli/install',
              ],
            },
            {
              type: 'group',
              text: 'Networks',
              link: '',
              children: [
                '/network',
                '/network/join-testnet',
                '/network/join-mainnet',
              ],
            },
            {
              type: 'group',
              text: 'Validating',
              link: '',
              children: [
                '/validators',
              ],
            },
            {
              type: 'group',
              text: 'Integrate',
              link: '',
              children: [
                '/integrate',
                '/integrate/token-listings',
              ],
            },
          ],


          // '/othernav-example': [
          //   {
          //     type: 'link-group',
          //     text: 'Other nav',
          //     link: '',
          //     children: [
          //     '/othernav',
          //     '/othernav/othernav',
          //     ],
          //   },
          // ],
        },
      }
    },

    themePlugins: {
      git: true,
    },
  },

  themeConfigOLd: {
    displayAllHeaders: true,
    logo: 'OSMOLogoTitleDark.png',
    logoDark: 'OSMOLogoTitleLight.png',
    repo: '',
    editLinks: false,
    docsDir: 'docs',
    editLinkText: '',
    lastUpdated: false,
    navbar: [
      // NavbarItem
      {
        text: 'Home',
        link: '/',
      },
      // NavbarGroup
      {
        text: 'Network',
        children: ['/network', '/validators'],
      },
    ],
    sidebarDepth: 0,

  },

  /**
   * Apply plugins，ref：https://v1.vuepress.vuejs.org/zh/plugin/
   */
  plugins: [
    '@vuepress/plugin-back-to-top',
    '@vuepress/plugin-medium-zoom',
    '@vuepress/plugin-docsearch',
    {
      locales: {
        '/': {
          placeholder: 'Search',
        },
      },
    },
  ]
})


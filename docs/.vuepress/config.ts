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
            text: 'Intro',
            link: '/intro',
          },
          {
            text: 'Develop',
            link: '/developing',
          },
          {
            text: 'Validate',
            link: '/validators/',
          },
          {
            text: 'Integrate',
            link: '/integrate/',
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
                '/intro/osmo',
                '/intro/terminology',
                '/intro/governance',
              ],
            },
            {
              type: 'group',
              text: 'Osmosis AMM App',
              link: '',
              children: [
                '/osmosis-app',
                '/osmosis-app/add-liquidity',
                '/osmosis-app/create-pool',
                '/osmosis-app/liquidity-bootstraping',
              ],
            },
            {
              type: 'group',
              text: 'Wallets',
              link: '',
              children: [
                '/wallets/',
              ],
            },



          ],

          '/developing': [
            {
              type: 'group',
              text: 'Command Line',
              link: '',
              children: [
                '/developing/cli',
                '/developing/cli/install',
              ],
            },
            {
              type: 'group',
              text: 'Networks',
              link: '',
              children: [
                '/developing/network',
                '/developing/network/join-testnet',
                '/developing/network/join-mainnet',
              ],
            },
          ],

          '/validators': [
            {
              type: 'group',
              text: 'Validate',
              link: '',
              children: [
                '/validators',
              ],
              collapsible: true,
            },
          ],

          '/integrate': [
            {
              type: 'group',
              text: 'Integrate',
              link: '',
              children: [
                '/integrate',
                '/integrate/token-listings',
              ],
              collapsible: true,
            },
              ],

          '/wallets/keplr': [
            {
              type: 'group',
              text: 'Keplr',
              link: '',
              children: [
                '/wallets/keplr/install-keplr',
                '/wallets/keplr/create-keplr-wallet',
                '/wallets/keplr/import-account',
                '/wallets/keplr/import-ledger-account',
              ],
              collapsible: true,
            }
          ],
        },
      }
    },

    themePlugins: {
      git: true,
    },
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


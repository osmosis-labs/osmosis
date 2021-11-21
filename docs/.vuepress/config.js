const { description } = require('../package')

module.exports = {
  /**
   * Ref：https://v1.vuepress.vuejs.org/config/#title
   */
  title: null,
  /**
   * Ref：https://v1.vuepress.vuejs.org/config/#description
   */
  description: description,
  /**
   * Extra tags to be injected to the page HTML `<head>`
   *
   * ref：https://v1.vuepress.vuejs.org/config/#head
   */
  head: [
    ['meta', { name: 'theme-color', content: '#803eaf' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }]
  ],
  /**
   * The base URL when deployed to Github Pages
   *
   * ref：https://vuepress.vuejs.org/config/#basic-config
   */
  base: "/osmosis/",

  /**
   * Theme configuration, here is the default theme configuration for VuePress.
   *
   * ref：https://v1.vuepress.vuejs.org/theme/default-theme-config.html
   */
  themeConfig: {
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
    sidebar: [
      {
        text: 'About',
        children: [
          '/intro',
          '/intro/terminology',
          '/osmo',
          '/governance',
          '/other-features',
        ],
      },
      {
        text: 'Wallets',
        children: [
          '/wallets',
        ],
      },
      {
        text: 'Liquidity',
        children: [
          '/liquidity',
          '/liquidity/liquidity-bootstraping',
        ],
      },
      {
        text: 'Command Line',
        children: [
          '/cli',
          '/cli/install',
        ],
      },
      {
        text: 'Networks',
        children: [
          '/network',
          '/network/join-testnet',
          '/network/join-mainnet',
        ],
      },
      {
        text: 'Validating',
        children: [
          '/validators',
        ],
      },
      {
        text: 'Integrate',
        children: [
          '/integrate',
          '/integrate/token-listings',
        ],
      },
    ],
  },

  /**
   * Apply plugins，ref：https://v1.vuepress.vuejs.org/zh/plugin/
   */
  plugins: [
    '@vuepress/plugin-back-to-top',
    '@vuepress/plugin-medium-zoom',
  ]
};

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
    docsDir: '',
    editLinkText: '',
    lastUpdated: false,
    nav: [
      {
        text: 'Home',
        link: '/',
      },

    ],
    sidebarDepth: 1,
    sidebar: [
      '/intro/',
      '/liquidity/',
      '/staking/',
      '/governance/',
      '/other-features/',
      '/osmo/',
      '/network/',
      '/validators/',
      '/tutorials/',
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

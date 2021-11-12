const { description } = require('../../package')

module.exports = {
  /**
   * Ref：https://v1.vuepress.vuejs.org/config/#title
   */
  title: 'Osmosis',
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
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }]
  ],

  /**
   * Theme configuration, here is the default theme configuration for VuePress.
   *
   * ref：https://v1.vuepress.vuejs.org/theme/default-theme-config.html
   */
  themeConfig: {
    displayAllHeaders: true,
    repo: '',
    editLinks: false,
    docsDir: '',
    editLinkText: '',
    lastUpdated: false,
    nav: [
      {
        text: 'Intro',
        link: '/intro/',
      },
      {
        text: 'Config',
        link: '/config/'
      }
    ],

    sidebar: [
      {
        path: '/intro/',      // optional, link of the title, which should be an absolute path and must exist
        collapsable: false, // optional, defaults to true
        sidebarDepth: 1,    // optional, defaults to 1
        children: [
          '/intro/'
        ]
      },
      {
        collapsable: false,
        children: [
          '/liquidity/',
        ],
        //initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/staking/',
        ],
        //initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/governance/',
        ],
        //initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/other-features/',
        ],
        //initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/network/',
        ],
       // initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/validators/',
        ],
        // initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      },
      {
        collapsable: false,
        children: [
          '/tutorials/',
        ],
        // initialOpenGroupIndex: -1 // optional, defaults to 0, defines the index of initially opened subgroup
      }
    ],

    sidebarworks: [
      '/intro/',
      '/network/',
    ],
    sidebarold: {
      auto: false,
      '/': [
        {
          title: 'Introduction',
          collapsable: false,
          children: [
            '/intro/',
            '/intro/getting-started',
          ]
        },
        {
          title: 'Networks',
          collapsable: false,
          auto: false,
          children: [
            '/network/osmosisd-setup',
            '/network/mainnet-setup',
            // '/network/joining-mainnet',
            // '/network/joining-testnet',
            // '/network/cosmovisor',
            // '/network/relayers',
          ]
        }
      ],

    }
  },

  /**
   * Apply plugins，ref：https://v1.vuepress.vuejs.org/zh/plugin/
   */
  plugins: [
    '@vuepress/plugin-back-to-top',
    '@vuepress/plugin-medium-zoom',
  ]
};

function getSideBar(folder, title) {
  const extension = [".md"];

  const files = fs
      .readdirSync(path.join(`${__dirname}/../${folder}`))
      .filter(
          (item) =>
              item.toLowerCase() != "readme.md" &&
              fs.statSync(path.join(`${__dirname}/../${folder}`, item)).isFile() &&
              extension.includes(path.extname(item))
      );

  return [{ title: title, children: ["", ...files] }];
}

module.exports = {
  title: "Osmosis Docs",
  markdown: {
    lineNumbers: true,
    extendMarkdown: (md) => {
      md.use(require("markdown-it-footnote"));
    },
  },
  base: "/osmosis/",
  description:
    "Osmosis - The Cosmos Interchain AMM",
  plugins: [
    [
      "@vuepress/register-components",
      {
        componentsDir: "theme/components",
      },
    ],
    [
      "vuepress-plugin-mathjax",
      {
        target: "svg",
        macros: {
          "*": "\\times",
        },
      },
    ],
  //  https://github.com/znicholasbrown/vuepress-plugin-code-copy
  //  ["vuepress-plugin-code-copy", {
  //    color: "#ffffff",
  //    backgroundColor: "#3e3383",
  //    }
  //  ],
    ["@maginapp/vuepress-plugin-copy-code", {
      color: "#ffffff",
      backgroundColor: "#ffffff",
      align: { bottom: '7px', right: '12px' },
      successText: " ",
      duration: 350,
    }
    ],
    [ 'tabs' ],
  ],
  head: [
    [
      "link",
      {
        rel: "stylesheet",
        type: "text/css",
        href: "https://cloud.typography.com/7420256/6416592/css/fonts.css",
      },
    ],
    [
      "link",
      {
        rel: "stylesheet",
        type: "text/css",
        href:
          "https://fonts.googleapis.com/css?family=Material+Icons|Material+Icons+Outlined",
      },
    ],

    [
      "link",
      {
        rel: "stylesheet",
        type: "text/css",
        href:
          "https://fonts.googleapis.com/css?family=Noto+Sans+KR:400,500,700&display=swap",
      },
    ],
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        href: "/img/favicon.png",
      },
    ],
    [
      "script",
      {},
      `window.onload = function() {
        requestAnimationFrame(function() {
          if (location.hash) {
            const element = document.getElementById(location.hash.slice(1))

            if (element) {
              element.scrollIntoView()
            }
          }
        })
      }`,
    ],
  ],
  themeConfig: {
    sidebarDepth: 3,
    smoothScroll: true,
    // overrideTheme: 'dark',
    // prefersTheme: 'dark',
    // overrideTheme: { light: [6, 18], dark: [18, 6] },
    // theme: 'default-prefers-color-scheme',
    logo: "/img/osmosis-logo-dark.svg",
    logoDark: "/img/osmosis-logo-light.svg",
    lastUpdated: "Updated on",
    repo: "osmosis-labs/osmosis",
    editLinks: true,
    editLinkText: "Edit this page on GitHub",
    docsBranch: 'main',
    docsDir: "docs",
    algolia: {
      apiKey: "PENDING", //TODO GET KEY
      indexName: "osmosis-project",
    },
    nav: [

      { text: 'Home', link: '/', },
      { text: 'Overview', link: '/overview/',},
      { text: 'Develop', link: '/developing/',},
      { text: 'Validate', link: '/validators/',},
      { text: 'Integrate', link: '/integrate/',},
      { text: 'Chat', link: 'https://v2.vuepress.vuejs.org/',},
      {
        text: "GitHub",
        link: "https://github.com/osmosis-labs/osmosis",
        icon: "/osmosis/img/github.svg",
      },
    ],
    sidebar: {
      "/overview/": [
        {
          title: "About",
          children: [
            '/overview/',
            '/overview/osmo',
            '/overview/terminology',
            '/overview/governance',
          ],
          collapsable: true,
        },
        {
          title: "Osmosis AMM App",
          children: [
            '/overview/osmosis-app/',
            '/overview/osmosis-app/learn-more',
          ],
          collapsable: true,
        },
        {
          title: 'Wallets',
          children: [
            '/overview/wallets/keplr/install-keplr',
            '/overview/wallets/keplr/create-keplr-wallet',
            '/overview/wallets/keplr/import-account',
            '/overview/wallets/keplr/import-ledger-account',
          ],
          collapsable: true,
        },
      ],


      '/developing': [
        {
          title: 'Home',
          children: [
            '/developing/',
          ],
          collapsable: false,
        },
        {
          title: 'Developer Guide',
          children: [
            '/developing/dev-guide',
            '/developing/cli/',
            '/developing/cli/install',
          ],
          collapsable: true,
        },
        {
          title: "osmosisd",
          children: [
            "/developing/osmosisd/",
            "/developing/osmosisd/commands",
            "/developing/osmosisd/subcommands",
          ],
          collapsable: true,
        },
        {
          title: 'Networks',
          children: [
            '/developing/network/join-testnet',
            '/developing/network/join-mainnet',
          ],
          collapsable: true,
        },
        {
          title: 'Modules',
          children: [
            "/developing/modules/",
            "/developing/modules/spec-claim",
            "/developing/modules/spec-epochs",
            "/developing/modules/spec-gamm",
            "/developing/modules/spec-gov",
            "/developing/modules/spec-lockup",
            "/developing/modules/spec-mint",
            "/developing/modules/spec-pool-incentives",
            "/developing/modules/spec-simulation",
            "/developing/modules/spec-txfees"
          ],
          collapsable: true,
        },
      ],
      '/validators': [
        {
          title: 'Home',
          children: [
            '/validators/',
          ],
          collapsable: false,
        },
        {
          title: 'Validate',
          children: [
            '/validators/validating-testnet',
            '/validators/validating-mainnet',
          ],
          collapsable: true,
        },
      ],

      '/integrate': [
        {
          title: 'Home',
          children: [
            '/integrate/',
          ],
          collapsable: false,
        },
        {
          title: 'Integrate',
          children: [
            '/integrate/token-listings',
          ],
          collapsable: true,
        },
      ],
      "/": [
        {
          title: "Overview",
          children: [
            "/history-and-changes",
          ],
          collapsable: false,
        },
      ],
    },

  },
};

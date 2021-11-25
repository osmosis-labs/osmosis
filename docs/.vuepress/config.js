module.exports = {
  title: "Osmosis Docs",
  markdown: {
    lineNumbers: true,
    extendMarkdown: (md) => {
      md.use(require("markdown-it-footnote"));
    },
  },
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

    // https://github.com/znicholasbrown/vuepress-plugin-code-copy
    ["vuepress-plugin-code-copy", {
      color: "#de1dd1",
      backgroundColor: "#3e3383",
      }
    ]

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
        icon: "/img/github.svg",
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
            '/overview/osmosis-app/add-liquidity',
            '/overview/osmosis-app/create-pool',
            '/overview/osmosis-app/liquidity-bootstraping',
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
        },
      ],


      '/developing': [
        {
          title: 'Developer Guide',
          children: [
            '/developing/',
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
            '/developing/network/',
            '/developing/network/join-mainnet',
            '/developing/network/join-testnet',
          ],
          collapsable: true,
        },
        {
          title: 'Modules',
          children: [
            "/developing/modules/spec-auth",
            "/developing/modules/spec-authz",
            "/developing/modules/spec-bank",
            "/developing/modules/spec-capability",
            "/developing/modules/spec-distribution",
            "/developing/modules/spec-evidence",
            "/developing/modules/spec-feegrant",
            "/developing/modules/spec-governance",
            "/developing/modules/spec-market",
            "/developing/modules/spec-oracle",
            "/developing/modules/spec-slashing",
            "/developing/modules/spec-staking",
            "/developing/modules/spec-treasury",
            "/developing/modules/spec-wasm",
          ],
          collapsable: true,
        },
      ],


      '/validators': [
        {
          title: 'Validate',
          children: [
            '/validators/',
            '/validators/validating-testnet',
          ],
          collapsible: true,
        },
      ],


      '/integrate': [
        {
          title: 'Integrate',
          children: [
          //'/integrate',
            '/integrate/token-listings',
          ],
          collapsible: true,
        },
      ],


      "/": [
        {
          title: "Overview",
          children: [
            "/history-and-changes",
          //"/migration-guide",
          ],
          collapsable: false,
        },
      ],
    },

  },
};

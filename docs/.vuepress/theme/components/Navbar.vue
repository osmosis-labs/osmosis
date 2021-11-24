<template>
  <header class="navbar">
    <SidebarButton @toggle-sidebar="$emit('toggle-sidebar')" />

    <RouterLink :to="$localePath" class="home-link">
      <img
        v-if="$site.themeConfig.logo"
        class="logo"
        :src="$withBase($site.themeConfig.logo)"
        :alt="$siteTitle"
      />
      <span
        v-if="$siteTitle"
        ref="siteName"
        class="site-name"
        :class="{ 'can-hide': $site.themeConfig.logo }"
        >{{ $siteTitle }}</span
      >
    </RouterLink>

    <div
      class="links"
      :style="
        linksWrapMaxWidth
          ? {
              'max-width': linksWrapMaxWidth + 'px',
            }
          : {}
      "
    >
      <NavLinks class="can-hide" />
      <AlgoliaSearchBox v-if="isAlgoliaSearch" :options="algolia" />
      <SearchBox
        v-else-if="
          $site.themeConfig.search !== false &&
            $page.frontmatter.search !== false
        "
      />
    </div>
  </header>
</template>

<script>
import AlgoliaSearchBox from "@AlgoliaSearchBox";
import SearchBox from "@SearchBox";
import SidebarButton from "@theme/components/SidebarButton.vue";
import NavLinks from "@theme/components/NavLinks.vue";

export default {
  name: "Navbar",

  components: {
    SidebarButton,
    NavLinks,
    SearchBox,
    AlgoliaSearchBox,
  },

  data() {
    return {
      linksWrapMaxWidth: null,
    };
  },

  computed: {
    algolia() {
      return (
        this.$themeLocaleConfig.algolia || this.$site.themeConfig.algolia || {}
      );
    },

    isAlgoliaSearch() {
      return this.algolia && this.algolia.apiKey && this.algolia.indexName;
    },
  },

  mounted() {
    const MOBILE_DESKTOP_BREAKPOINT = 719; // refer to config.styl
    const NAVBAR_VERTICAL_PADDING =
      parseInt(css(this.$el, "paddingLeft")) +
      parseInt(css(this.$el, "paddingRight"));
    const handleLinksWrapWidth = () => {
      if (document.documentElement.clientWidth < MOBILE_DESKTOP_BREAKPOINT) {
        this.linksWrapMaxWidth = null;
      } else {
        this.linksWrapMaxWidth =
          this.$el.offsetWidth -
          NAVBAR_VERTICAL_PADDING -
          ((this.$refs.siteName && this.$refs.siteName.offsetWidth) || 0);
      }
    };
    handleLinksWrapWidth();
    window.addEventListener("resize", handleLinksWrapWidth, false);
  },
};

function css(el, property) {
  // NOTE: Known bug, will return 'auto' if style value is 'auto'
  const win = el.ownerDocument.defaultView;
  // null means not to return pseudo styles
  return win.getComputedStyle(el, null)[property];
}
</script>

<style lang="stylus">
$navbar-vertical-padding = 1.25rem;
$navbar-horizontal-padding = 2.5rem;
$mobile-navbar-vertical-padding = 0.75rem;
$sm-mobile-navbar-horizontal-padding = 1.5rem;

.navbar {
  padding: $navbar-vertical-padding $navbar-horizontal-padding;
  line-height: 3rem;
  width: 100%;
  max-width: $layoutWidth;
  margin: 0 auto;

  a, span, img {
    display: inline-block;
  }

  .logo {
    height: 3rem;
    width: auto;
    margin-right: 0.8rem;
    vertical-align: top;
  }

  .theme-dark .logo {
    filter: brightness(149%);
  }
  .site-name {
    font-size: 1.3rem;
    font-weight: 700;
    color: var(--text-color);
    position: relative;
    display: none;
  }

  .links {
    padding-left: 1.5rem;
    box-sizing: border-box;
    white-space: nowrap;
    font-size: 0.9rem;
    position: absolute;
    right: $navbar-horizontal-padding;
    top: $navbar-vertical-padding;
    display: flex;

    .search-box {
      flex: 0 0 auto;
      vertical-align: top;
      margin-left: 1.5rem;
      margin-right: 0 !important;
      display: flex;
      align-items: center;

      input {
        border: 0;
        font-size: 1rem;
        -webkit-transition: all 0.2s ease;
        transition: all 0.2s ease;
        height: 2.625rem;
        width: 2.625rem;
        border-radius: 1.3125rem;
        font-size: 1rem;
        padding: 0;
        background: rgba($primaryColor, 0.08) url('/img/search.svg') 0.75rem center no-repeat;
        background-size: 1.125rem;
        color: transparent;

        &:focus, &:active {
          background-color: #fff;
          box-shadow: 0px 0.125rem 0.375rem 0px rgba(0, 0, 0, 0.2);
          padding-left: 2.25rem;
          padding-right: 1rem;
          width: 10rem;
          color: $textColor;
        }
      }
    }
  }
}

@media (max-width: $MQMobile) {
  .theme-dark .navbar .logo {
      filter: brightness(149%);
  }
  .navbar {
    padding-top: $mobile-navbar-vertical-padding;
    padding-bottom: $mobile-navbar-vertical-padding;
    padding-right: 2rem;
    padding-left: 2rem;
    line-height: 1;

    .logo {
      height: 2.25rem;
    }

    .can-hide {
      display: none;
    }

    .site-name {
      width: calc(100vw - 9.4rem);
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    .links {
      position: absolute;
      padding: 0;
      right: 4.875rem;
      top: $mobile-navbar-vertical-padding;
        .search-box {
          padding-top: 0;
          input {
            height: 2.25rem;
            width: 2.25rem;
            border-radius: 1.125rem;
            background: rgba($primaryColor, 0.08) url('/img/search.svg') 0.625rem center no-repeat;
            left: 0;

            &:active, &:focus {
              padding-left: 2.25rem;
              padding-right: 1rem;
              width: calc(100vw - 6.5rem);
              box-sizing: border-box;
            }
          }
        }
    }
  }
}

@media (max-width: $MQMobileNarrow) {
  .navbar {
    padding-top: $mobile-navbar-vertical-padding;
    padding-right: 1.5rem;
    padding-left: 1.5rem;
    .links {
      right: 4.375rem;
      .search-box {
        input {
          &:active, &:focus {
            width: calc(100vw - 5.625rem);
          }
        }
      }
    }
  }
}
</style>

<template>
  <nav v-if="userLinks.length || repoLink" class="nav-links">
    <!-- user links -->
    <div v-for="item in userLinks" :key="item.link" class="nav-item">
      <DropdownLink v-if="item.type === 'links'" :item="item" />
      <NavLink v-else :item="item" />
    </div>

    <!-- repo link (disabled)
    <a
      v-if="repoLink"
      :href="repoLink"
      class="repo-link"
      target="_blank"
      rel="noopener noreferrer"
    >{{ repoLabel }}</a>-->
  </nav>
</template>

<script>
import DropdownLink from "@theme/components/DropdownLink.vue";
import { resolveNavLinkItem } from "../util";
import NavLink from "@theme/components/NavLink.vue";

export default {
  name: "NavLinks",

  components: {
    NavLink,
    DropdownLink
  },

  computed: {
    userNav() {
      return this.$themeLocaleConfig.nav || this.$site.themeConfig.nav || [];
    },

    nav() {
      const { locales } = this.$site;
      if (locales && Object.keys(locales).length > 1) {
        const currentLink = this.$page.path;
        const routes = this.$router.options.routes;
        const themeLocales = this.$site.themeConfig.locales || {};
        const languageDropdown = {
          text: this.$themeLocaleConfig.selectText || "Languages",
          ariaLabel: this.$themeLocaleConfig.ariaLabel || "Select language",
          items: Object.keys(locales).map(path => {
            const locale = locales[path];
            const text =
              (themeLocales[path] && themeLocales[path].label) || locale.lang;
            let link;
            // Stay on the current page
            if (locale.lang === this.$lang) {
              link = currentLink;
            } else {
              // Try to stay on the same page
              link = currentLink.replace(this.$localeConfig.path, path);
              // fallback to homepage
              if (!routes.some(route => route.path === link)) {
                link = path;
              }
            }
            return { text, link };
          })
        };
        return [...this.userNav, languageDropdown];
      }
      return this.userNav;
    },

    userLinks() {
      return (this.nav || []).map(link => {
        return Object.assign(resolveNavLinkItem(link), {
          items: (link.items || []).map(resolveNavLinkItem)
        });
      });
    },

    repoLink() {
      const { repo } = this.$site.themeConfig;
      if (repo) {
        return /^https?:/.test(repo) ? repo : `https://github.com/${repo}`;
      }
      return null;
    },

    repoLabel() {
      if (!this.repoLink) return;
      if (this.$site.themeConfig.repoLabel) {
        return this.$site.themeConfig.repoLabel;
      }

      const repoHost = this.repoLink.match(/^https?:\/\/[^/]+/)[0];
      const platforms = ["GitHub", "GitLab", "Bitbucket"];
      for (let i = 0; i < platforms.length; i++) {
        const platform = platforms[i];
        if (new RegExp(platform, "i").test(repoHost)) {
          return platform;
        }
      }

      return "Source";
    }
  }
};
</script>

<style lang="stylus">
.nav-links {
  display: flex;
  align-items: center;

  a {
    line-height: 1.4rem;
    color: var(--primary-color);
    text-transform: uppercase;
    font-size: 0.875rem;
    font-weight: 500;

    &:hover, &.router-link-active {
      color: var(--primary-color);
    }
  }

  .nav-item {
    position: relative;
    display: inline-block;
    margin-left: 2rem;
    line-height: 2rem;

    &:first-child {
      margin-left: 0;
    }
  }

  .repo-link {
    margin-left: 1.5rem;
  }

  .nav-icon {
    height: 2.625rem;
    width: 2.625rem;
    border-radius: 1.3125rem;
    padding: 0;
    background: var(--primary-color)
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;

    img {
      width: 1.125rem;
      height: 1.125rem;
    }

    &:hover {
      background: var(--primary-color);
    }
  }
}

@media (max-width: $MQMobile) {
  .nav-links {
    align-items: flex-end;

    a {
      line-height: 1.1;
      color: var(--primary-color);
      text-transform: uppercase;
      font-size: 0.9375rem;
      font-weight: 700;
      padding: 0 0.125rem;
      border-bottom: solid 3px transparent;
      height: 2.875rem;
      display: flex;
      align-items: center;

      &.router-link-active {
        color: $accentColor;
        border-bottom: solid 3px $accentColor;
      }
    }

    .nav-item, .repo-link {
      margin-left: 0;
    }

    .nav-icon {
      height: unset;
      width: unset;
      border-radius: 0;
      padding: 0 0.25rem;
      background: transparent;
      display: block;

      img {
        width: 1.125rem;
        height: 1.125rem;
      }

      &:hover {
        background: transparent;
      }
    }
  }
}

@media (min-width: $MQMobile) {
  .nav-links a {
    &:hover, &.router-link-active {
      color: var(--primary-color);
    }
  }

  .nav-item > a:not(.external) {
    &:hover, &.router-link-active {
      margin-bottom: -2px;
      border-bottom: 2px solid lighten($accentColor, 8%);
    }
  }
}
</style>

<template>
  <nav v-if="bread.length > 1" id="breadcrumb" class="breadcrumb-container">
    <router-link class="breadcrumb" v-for="crumb in bread" :key="crumb.path" :to="crumb.path">
      <span>{{crumb.title}}</span>
    </router-link>
  </nav>
</template>

<script>
export default {
  name: "Breadcrumb",
  computed: {
    bread() {
      const parts = this.$page.path.split("/");
      if (!parts[parts.length - 1].length) {
        parts.pop();
      }
      let link = "";
      const crumbs = [];
      for (let i = 0; i < parts.length; i++) {
        link += parts[i];
        const page = this.$site.pages.find(
          el => el.path === link || el.path === link + "/"
        );
        link += "/";
        if (page != null) {
          crumbs.push({
            path: page.path,
            title: page.frontmatter.breadcrumb || page.title
          });
        }
      }
      return crumbs;
    }
  }
};
</script>

<style lang="stylus" scoped>
.breadcrumb-container {
  margin-top: ($navbarHeight - 0.25rem);
  margin-bottom: - $navbarHeight;
  padding: 0 2.5rem 1rem;
}

.breadcrumb {
  font-size: 0.75rem;
  text-transform: uppercase;
  font-weight: 500;
  color: var(--text-color);
  line-height: 1.8;

  &:first-child {
    span {
      font-size: 0;

      &:before {
        content: 'Docs';
        font-size: 0.75rem;
      }
    }

    &:before {
      content: '';
      margin: 0;
    }
  }

  &::before {
    content: ' // ';
    color: var(--text-color);
    font-weight: 400;
    margin: 0 0.5rem;
  }

  &:last-child {
    cursor: default;
    color: var(--primary-color);
  }
}

@media (max-width: $MQMobile) {
  .breadcrumb-container {
    padding: 0 2rem 2rem;
    margin-top: 3.75rem;
  }
}

@media (max-width: $MQMobileNarrow) {
  .breadcrumb-container {
    padding: 0 1.5rem 2rem;
  }
}
</style>

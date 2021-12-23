<script>
import { isActive, hashRE, groupHeaders } from "../util";

export default {
  functional: true,

  props: ["item", "sidebarDepth"],

  render(
    h,
    {
      parent: { $page, $site, $route, $themeConfig, $themeLocaleConfig },
      props: { item, sidebarDepth }
    }
  ) {
    // use custom active class matching logic
    // due to edge case of paths ending with / + hash
    const selfActive = isActive($route, item.path);
    // for sidebar: auto pages, a hash link should be active if one of its child
    // matches
    const active =
      item.type === "auto"
        ? selfActive ||
          item.children.some(c =>
            isActive($route, item.basePath + "#" + c.slug)
          )
        : selfActive;
    const link =
      item.type === "external"
        ? renderExternal(h, item.path, item.title || item.path)
        : renderLink(h, item.path, item.title || item.path, active);

    const maxDepth = [
      $page.frontmatter.sidebarDepth,
      sidebarDepth,
      $themeLocaleConfig.sidebarDepth,
      $themeConfig.sidebarDepth,
      1
    ].find(depth => depth !== undefined);

    const displayAllHeaders =
      $themeLocaleConfig.displayAllHeaders || $themeConfig.displayAllHeaders;

    if (item.type === "auto") {
      return [
        link,
        renderChildren(h, item.children, item.basePath, $route, maxDepth)
      ];
    } else if (
      (active || displayAllHeaders) &&
      item.headers &&
      !hashRE.test(item.path)
    ) {
      const children = groupHeaders(item.headers);
      return [link, renderChildren(h, children, item.path, $route, maxDepth)];
    } else {
      return link;
    }
  }
};

function renderLink(h, to, text, active, level) {
  const component = {
    props: {
      to,
      activeClass: "",
      exactActiveClass: ""
    },
    class: {
      active,
      "sidebar-link": true
    }
  };

  if (level > 2) {
    component.style = {
      "padding-left": level + "rem"
    };
  }

  return h("RouterLink", component, text);
}

function renderChildren(h, children, path, route, maxDepth, depth = 1) {
  if (!children || depth > maxDepth) return null;
  return h(
    "ul",
    { class: "sidebar-sub-headers" },
    children.map(c => {
      const active = isActive(route, path + "#" + c.slug);
      return h("li", { class: "sidebar-sub-header" }, [
        renderLink(h, path + "#" + c.slug, c.title, active, c.level - 1),
        renderChildren(h, c.children, path, route, maxDepth, depth + 1)
      ]);
    })
  );
}

function renderExternal(h, to, text) {
  return h(
    "a",
    {
      attrs: {
        href: to,
        target: "_blank",
        rel: "noopener noreferrer"
      },
      class: {
        "sidebar-link": true
      }
    },
    [text]
  );
}
</script>

<style lang="stylus">
.sidebar .sidebar-sub-header {
  a {
    color: var(--primary-color);

    &:before {
      content: '';
      display: inline-block;
      width: 12px;
      height: 12px;
      background: url('/img/bullet_osmo_gray.svg');
      background-size: 12px 12px;
      margin-right: 0.625rem;
      vertical-align: middle;
      margin-top: -3px;
    }

    &.active {
      &:before {
        content: '';
        display: inline-block;
        width: 12px;
        height: 12px;
        background: url('/img/bullet_osmo.svg');
        background-size: 12px 12px;
        margin-right: 0.625rem;
      }
    }
  }
}

a.sidebar-link {
  font-weight: 400;
  display: inline-block;
  color: var(--primary-color);
  line-height: 1.25rem;
  padding: 0.313rem 0;
  width: 100%;
  box-sizing: border-box;
  letter-spacing: -0.006rem;

  &:hover {
    color: $accentColor;
  }

  &.active {
    font-weight: 500;
    color: var(--primary-color);
    border-left-color: $accentColor;
  }

  .sidebar-sub-headers & {
    border-left: none;

    &.active {
      font-weight: 500;
    }
  }
}

@media (max-width: $MQMobile) {
  a.sidebar-link {
    padding: 0.375rem 0;
  }
}
</style>

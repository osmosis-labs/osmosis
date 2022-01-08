<template>
  <div v-if="prev || next" class="page-nav">
    <div class="inner">
      <div v-if="prev" class="prev">
        <a
          v-if="prev.type === 'external'"
          :href="prev.path"
          target="_blank"
          rel="noopener noreferrer"
        >{{ prev.title || prev.path }}</a>

        <RouterLink v-else :to="prev.path">{{ prev.title || prev.path }}</RouterLink>
      </div>
      <div v-if="next" class="next">
        <a
          v-if="next.type === 'external'"
          :href="next.path"
          target="_blank"
          rel="noopener noreferrer"
        >{{ next.title || next.path }}</a>

        <RouterLink v-else :to="next.path">{{ next.title || next.path }}</RouterLink>
      </div>
    </div>
  </div>
</template>

<script>
import { resolvePage } from "../util";
import isString from "lodash/isString";
import isNil from "lodash/isNil";

export default {
  name: "PageNav",

  props: ["sidebarItems"],

  computed: {
    prev() {
      return resolvePageLink(LINK_TYPES.PREV, this);
    },

    next() {
      return resolvePageLink(LINK_TYPES.NEXT, this);
    }
  }
};

function resolvePrev(page, items) {
  return find(page, items, -1);
}

function resolveNext(page, items) {
  return find(page, items, 1);
}

const LINK_TYPES = {
  NEXT: {
    resolveLink: resolveNext,
    getThemeLinkConfig: ({ nextLinks }) => nextLinks,
    getPageLinkConfig: ({ frontmatter }) => frontmatter.next
  },
  PREV: {
    resolveLink: resolvePrev,
    getThemeLinkConfig: ({ prevLinks }) => prevLinks,
    getPageLinkConfig: ({ frontmatter }) => frontmatter.prev
  }
};

function resolvePageLink(
  linkType,
  { $themeConfig, $page, $route, $site, sidebarItems }
) {
  const { resolveLink, getThemeLinkConfig, getPageLinkConfig } = linkType;

  // Get link config from theme
  const themeLinkConfig = getThemeLinkConfig($themeConfig);

  // Get link config from current page
  const pageLinkConfig = getPageLinkConfig($page);

  // Page link config will overwrite global theme link config if defined
  const link = isNil(pageLinkConfig) ? themeLinkConfig : pageLinkConfig;

  if (link === false) {
    return;
  } else if (isString(link)) {
    return resolvePage($site.pages, link, $route.path);
  } else {
    return resolveLink($page, sidebarItems);
  }
}

function find(page, items, offset) {
  const res = [];
  flatten(items, res);
  for (let i = 0; i < res.length; i++) {
    const cur = res[i];
    if (cur.type === "page" && cur.path === decodeURIComponent(page.path)) {
      return res[i + offset];
    }
  }
}

function flatten(items, res) {
  for (let i = 0, l = items.length; i < l; i++) {
    if (items[i].type === "group") {
      flatten(items[i].children || [], res);
    } else {
      res.push(items[i]);
    }
  }
}
</script>

<style lang="stylus">
@require '../styles/wrapper.styl';

.page-nav {
  @extend $wrapper;
  padding-top: 1rem;
  padding-bottom: 0;

  .inner {
    width: auto;
    display: -webkit-box;
    display: flex;
    -webkit-box-pack: justify;
    justify-content: space-between;
    margin-left: -12px;
    margin-right: -12px;
  }

  .prev, .next {
    flex: 1;
    a {
      display: block;
      font-size: 16px;
      border-radius: 35px;
      padding: 33px 30px 16px;
      margin: 0 12px;
      position: relative;
      transition: 0.2s;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      border: 1px solid var(--primary-color);
      color: var(--primary-color);
      font-weight: 500;
      line-height: 1.2em;
      box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);

      &:hover {
        box-shadow: 0px 10px 20px 0px rgba(0, 0, 0, 0.1);
      }

      &:before {
        content: 'PREVIOUS';
        position: absolute;
        top: 14px;
        right: 30px;
        color: var(--primary-color);
        opacity: 0.5;
        font-size: 10px;
        font-weight: 700;
      }

      &:after {
        content: '';
        display: block;
        position: absolute;
        top: 30px;
        left: 25px;
        padding: 5px;
        box-shadow: 2px -2px $primaryColor inset;
        border: 0 solid transparent;
        transition: 0.2s;
        transform: rotate(45deg);
      }
    }
  }

  .prev {
    a {
      margin-top: 15px;
      text-align: right;
      padding-left: 40px;
    }
  }

  .next {
    a {
      margin-top: 15px;
      text-align: left;
      padding-right: 40px;

      &:before {
        content: 'NEXT';
        right: unset;
        left: 30px;
      }

      &:after {
        left: unset;
        right: 25px;
        transform: rotate(225deg);
      }
    }
  }
}

@media (max-width: $MQMobile) {
  .page-nav .inner {
    flex-direction: column-reverse;
  }
}
</style>

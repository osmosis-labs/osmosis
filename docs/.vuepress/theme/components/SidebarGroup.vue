<template>
  <section
    class="sidebar-group"
    :class="[
      {
        collapsable,
        'is-sub-group': depth !== 0
      },
      `depth-${depth}`
    ]"
  >
    <RouterLink
      v-if="item.path"
      class="sidebar-heading clickable"
      :class="{
        open,
        active: isActive($route, item.path)
      }"
      :to="item.path"
      @click.native="$emit('toggle')"
    >
      <span>{{ item.title }}</span>
      <span v-if="collapsable" class="arrow" :class="open ? 'down' : 'right'" />
    </RouterLink>

    <p v-else class="sidebar-heading" :class="{ open }" @click="$emit('toggle')">
      <span>{{ item.title }}</span>
      <span v-if="collapsable" class="arrow" :class="open ? 'down' : 'right'" />
    </p>

    <DropdownTransition>
      <SidebarLinks
        v-if="open || !collapsable"
        class="sidebar-group-items"
        :items="item.children"
        :sidebar-depth="item.sidebarDepth"
        :depth="depth + 1"
      />
    </DropdownTransition>
  </section>
</template>

<script>
import { isActive } from "../util";
import DropdownTransition from "@theme/components/DropdownTransition.vue";

export default {
  name: "SidebarGroup",

  components: {
    DropdownTransition
  },

  props: ["item", "open", "collapsable", "depth"],

  // ref: https://vuejs.org/v2/guide/components-edge-cases.html#Circular-References-Between-Components
  beforeCreate() {
    this.$options.components.SidebarLinks = require("@theme/components/SidebarLinks.vue").default;
  },

  methods: { isActive }
};
</script>

<style lang="stylus">
.sidebar-group {
  .sidebar-group {
    padding-left: 0.5em;
  }

  &:not(.collapsable) {
    .sidebar-heading:not(.clickable) {
      cursor: auto;
    }
  }

  // refine styles of nested sidebar groups
  &.is-sub-group {
    padding-left: 0;

    & > .sidebar-heading {
      font-size: 0.95em;
      line-height: 1.4;
      font-weight: normal;
      padding-left: 2rem;

      &:not(.clickable) {
        opacity: 0.5;
      }
    }

    & > .sidebar-group-items {
      padding-left: 1rem;

      & > li > .sidebar-link {
        font-size: 0.95em;
        border-left: none;
      }
    }
  }

  &.depth-2 {
    & > .sidebar-heading {
      border-left: none;
    }
  }
}

.sidebar-heading {
  color: var(--primary-color);
  transition: color 0.15s ease;
  cursor: pointer;
  font-size: 0.813rem;
  font-weight: 500;
  text-transform: uppercase;
  width: 100%;
  box-sizing: border-box;
  margin: 0 0 0.313rem;

  &.open, &:hover {
    color: var(--primary-color);
  }

  .arrow {
    position: relative;
    top: -1px;
    left: 0.5em;

    &.right {
      top: 0;
    }
  }

  &.clickable {
    &.active {
      font-weight: 500;
      color: $primaryColor;
    }

    &:hover {
      color: $accentColor;
    }
  }
}

.sidebar-group-items {
  transition: height 0.1s ease-out;
  font-size: 0.95em;
  overflow: hidden;

  a.sidebar-link {
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

  a.sidebar-link[target=_blank] {
    &:before {
      content: '';
      display: inline-block;
      width: 12px;
      height: 12px;
      background: url('/img/external.svg');
      background-size: 12px 12px;
      margin-right: 0.625rem;
    }
  }

  .sidebar-sub-headers {
    padding-left: 5px;

    .sidebar-sub-header {
      a {
        border-left: solid 1px darken($borderColor, 5%);
        padding-left: 1.563rem;

        &:before {
          display: none;
        }

        &.active {
          border-left: solid 1px $primaryColor;

          &:before {
            display: none;
          }
        }
      }
    }

    .sidebar-sub-headers {
      padding-left: 0;

      .sidebar-sub-header {
        a {
          border-left: solid 1px darken($borderColor, 5%);
          padding-left: 2.188rem;
          font-size: 0.78rem;
          color: #777;

          &:before {
            display: none;
          }

          &.active {
            border-left: solid 1px $primaryColor;
            color: $primaryColor;
            font-weight: 500;

            &:before {
              display: none;
            }
          }
        }
      }
    }
  }
}

@media (max-width: $MQMobile) {
  .sidebar-heading {
    font-size: 1rem;
  }
}
</style>

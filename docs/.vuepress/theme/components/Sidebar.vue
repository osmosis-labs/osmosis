<template>
  <aside class="sidebar">
    <NavLinks />

    <slot name="top" />

    <SidebarLinks :depth="0" :items="items" />
    <slot name="bottom" />
    <div class="sidebar_footer">
    </div>
  </aside>
</template>

<script>
import SidebarLinks from "@theme/components/SidebarLinks.vue";
import NavLinks from "@theme/components/NavLinks.vue";

export default {
  name: "Sidebar",

  components: { SidebarLinks, NavLinks },

  props: ["items"]
};
</script>

<style lang="stylus">
.sidebar {
  ul {
    padding: 0;
    margin: 0;
    list-style-type: none;
  }

  a {
    display: inline-block;
  }

  .nav-links {
    display: none;
    border-bottom: 1px solid $borderColor;
    padding: 0.5rem 2.5rem 0.75rem;

    a {
      font-weight: 500;
    }

    .nav-item, .repo-link {
      display: block;
      line-height: 1.25rem;
      font-size: 1.1em;
      padding: 0.5rem 0 0.5rem 1.5rem;
    }
  }

  & > .sidebar-links {
    padding: 1.25rem 2.5rem 1.5rem;

    & > li > a.sidebar-link {
      font-size: 0.813rem;
      text-transform: uppercase;
      line-height: 1.7;
      padding: 0;
      font-weight: 500;
      margin-bottom: 0.313rem;
      word-break: break-all;
      overflow: visible;
      white-space: normal;
    }

    & > li:not(:first-child) {
      margin-top: 0.938rem;
    }
  }

  .sidebar_footer {
    padding: 0.5rem 2.5rem;
    background-color: var(--background-color);
    position: sticky;
    bottom: 0;

    &:before {
      content: '';
      display: block;
      background-color: var(--background-color);
      height: 1px;
      margin-top: -0.5rem;
      margin-bottom: 1.25rem;
    }

    a {
      display: block;
      position: relative;
      color: var(--primary-color);
      margin: 0.75rem 0;
      padding-left: 2.875rem;
      font-weight: 400;
      font-size: 0.875rem;
      line-height: 1.1;

      &:hover {
        .image {
          background: rgba($primaryColor, 0.15);
        }
      }

      &:first-child {
      }

      &:last-child {
      }

      i {
        font-size: 1rem;
        vertical-align: middle;
        margin-top: -1px;
        transform: rotate(180deg);
      }

      .name {
        display: flex;
        flex-direction: column;
        height: 2.25rem;
        justify-content: center;
      }

      small {
        font-size: 0.6875rem;
      }

      .image {
        position: absolute;
        width: 2.25rem;
        height: 2.25rem;
        display: flex;
        align-items: center;
        justify-content: center;
        background: $borderColor;
        border-radius: 50%;
        box-sizing: border-box;
        line-height: 1;
        top: 0;
        left: 0;
        transition: 0.2s;

        img {
          height: 1.5rem;
          width: auto;
        }
      }

      &.comingsoon {
        cursor: default;
        filter: grayscale(100%);
        opacity: 0.75;
      }
    }
  }
}

@media (max-width: $MQMobile) {
  .sidebar {
    right: 0;
    height: 100vh;

    .nav-links {
      display: flex;
      padding: 0 2rem;
      justify-content: space-between;
      overflow-x: auto;

      a {
        font-weight: 700;
      }

      .nav-item {
        padding: 0;
        text-align: center;
        padding-right: 20px;
        white-space: nowrap;

        a {
          display: flex;
          justify-content: center;
        }
      }

      .dropdown-wrapper .nav-dropdown .dropdown-item a.router-link-active::after {
        top: calc(1rem - 2px);
      }
    }

    & > .sidebar-links {
      padding: 2rem;

      & > li > a.sidebar-link {
        font-size: 1rem;
      }
    }
  }
}

@media (max-width: $MQMobileNarrow) {
  .sidebar {
    .nav-links {
      padding: 0 1.5rem;
    }

    & > .sidebar-links {
      padding: 1.5rem;
    }
  }
}
</style>

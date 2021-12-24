<template>
  <RouterLink
    v-if="isInternal && this.$page.path.split('/').length < 3"
    class="nav-link"
    :to="link"
    @focusout.native="focusoutAction"
  >{{ item.text }}</RouterLink>
  <RouterLink
    v-else-if="isInternal"
    class="nav-link"
    :to="link"
    :exact="exact"
    @focusout.native="focusoutAction"
  >{{ item.text }}</RouterLink>
  <a
    v-else
    :href="link"
    class="nav-link external"
    :target="target"
    :rel="rel"
    @focusout="focusoutAction"
  >
    <div v-if="item.icon" class="nav-icon"><img :src="item.icon"></div>
    <div v-else>{{ item.text }}</div>
  </a>
</template>

<script>
import { isExternal, isMailto, isTel, ensureExt } from "../util";

export default {
  name: "NavLink",

  props: {
    item: {
      required: true
    }
  },

  computed: {
    link() {
      return ensureExt(this.item.link);
    },

    exact() {
      if (this.$site.locales) {
        return Object.keys(this.$site.locales).some(
          rootLink => rootLink === this.link
        );
      }
      return this.link === "/";
    },

    isNonHttpURI() {
      return isMailto(this.link) || isTel(this.link);
    },

    isBlankTarget() {
      return this.target === "_blank";
    },

    isInternal() {
      return !isExternal(this.link) && !this.isBlankTarget;
    },

    target() {
      if (this.isNonHttpURI) {
        return null;
      }
      if (this.item.target) {
        return this.item.target;
      }
      return isExternal(this.link) ? "_blank" : "";
    },

    rel() {
      if (this.isNonHttpURI) {
        return null;
      }
      if (this.item.rel) {
        return this.item.rel;
      }
      return this.isBlankTarget ? "noopener noreferrer" : "";
    }
  },

  methods: {
    focusoutAction() {
      this.$emit("focusout");
    }
  }
};
</script>

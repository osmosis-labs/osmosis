<template>
  <span>
    <a v-if="link" :href="link">
      <code :class="classProp">{{to}}</code>
    </a>
    <code v-else :class="classProp">{{to}}</code>
  </span>
</template>

<script>
import symbols from "../../symbols.js";

export default {
  name: "Reference",
  props: ["to", "classProp"],
  computed: {
    link() {
      let matches = this.to.match(/([^\[\]]*)(\[.*\])?/);
      if (matches[1] in symbols) {
        return symbols[matches[1]];
      } else {
        return null;
      }
    }
  }
};
</script>

<style lang="stylus"></style>
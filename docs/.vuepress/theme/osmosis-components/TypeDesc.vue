<template>
  <tr class="type-desc">
    <td>
      <reference class="type-desc-type" :to="type" />
    </td>
    <td class="nopadding">
      <slot />
    </td>
  </tr>
</template>

<script>
export default {
  name: "TypeDesc",
  props: ["type"]
};
</script>

<style lang="stylus">
.type-desc {
  border: none;

  td {
    border: none;
  }

  code.type-desc-type {
    background: none;
    padding: 0;
    font-weight: bold;
    line-height: 1.7;
  }
}

.nopadding > p {
  display: inline !important;
}
</style>
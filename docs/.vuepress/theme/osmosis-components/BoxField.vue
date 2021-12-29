<template>
  <tr class="box-field">
    <td valign="top">
      <reference classProp="box-field-type" :to="type" />
      <br />
      <b>{{ name }}</b>
    </td>
    <td valign="top" class="nopadding">
      <br />
      <slot />
    </td>
  </tr>
</template>

<script>
export default {
  name: "BoxField",
  props: ["name", "type"]
};
</script>

<style lang="stylus">
.box-field {
  border: none;

  td {
    border: none;
  }

  code.box-field-type {
    background: none;
    padding: 0;
    font-size: 0.75rem;
    font-weight: bold;
    line-height: 1.7;
  }
}

.nopadding > p {
  display: inline !important;
}
</style>
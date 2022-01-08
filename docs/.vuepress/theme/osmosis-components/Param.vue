<template>
  <tr class="param">
    <td>
      <reference classProp="param-type" :to="type" />
      <br />
      <b v-if="this.optional === undefined">{{ name }}</b>
      <em v-else>{{ name }}</em>
      <div v-if="this.optional !== undefined">
        <em class="param-optional">optional</em>
      </div>
    </td>
    <td class="nopadding">
      <slot />
      <br />
      <div v-if="this.default !== undefined">
        <b>DEFAULT</b>
        <code>{{ this.default }}</code>
      </div>
    </td>
  </tr>
</template>

<script>
export default {
  name: "Param",
  props: ["name", "type", "default", "optional", "default"]
};
</script>

<style lang="stylus">
.param {
  border: none;

  td {
    border: none;
  }

  code.param-type {
    background: none;
    padding: 0;
    font-size: 0.75rem;
    font-weight: bold;
    line-height: 1.5;
  }

  em {
    font-weight: 500;
  }

  em.param-optional {
    font-size: 0.625rem;
    color: #444;
    font-weight: 400;
  }
}

.nopadding > p {
  display: inline !important;
}
</style>
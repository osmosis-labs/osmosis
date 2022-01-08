<template>
  <div>
    <component
      v-if="dynamicComponent"
      :is="dynamicComponent"
      :rootKey="root"
      :data="data"
      class="osmo-json"
    />
  </div>
</template>

<script>
export default {
  name: "JsonTreeView",
  props: ["root", "data"],
  data() {
    return {
      dynamicComponent: null
    };
  },
  mounted() {
    import("vue-json-component").then(module => {
      this.dynamicComponent = module.JSONView;
    });
  }
};
</script>

<style lang="scss">
.osmo-json {
  --vjc-key-color: #000000 !important;
  --vjc-valueKey-color: #525252 !important;
  --vjc-string-color: #3a74e0 !important;
  --vjc-number-color: #2aa198;
  --vjc-boolean-color: #cb4b16;
  --vjc-null-color: #6c71c4;
  --vjc-arrow-size: 5px !important;
  --vjc-arrow-color: #444;
  --vjc-hover-color: rgba(0, 0, 0, 0) !important;
  font-family: "JetBrainsMono";
}
</style>

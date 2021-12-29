<template>
  <div>
    <table class="param-list" v-if="noindent === undefined">
      <!-- <thead>
      <th>Name</th>
      <th>Type</th>
      <th>Default</th>
      <th style="width: 100%"Description</th>
    </thead>>
      -->
      <slot />
    </table>
    <table class="param-list noindent" v-else>
      <!-- <thead>
      <th>Name</th>
      <th>Type</th>
      <th>Default</th>
      <th style="width: 100%"Description</th>
    </thead>>
      -->
      <slot />
    </table>
  </div>
</template>

<script>
export default {
  name: "ParamsList",
  props: ["params", "noindent"]
};
</script>

<style lang="stylus">
.indent {
  .param-list {
    margin-top: -1rem;
    .param-list {
      margin-top: 0;
    }
  }
}

.param-list {
  width: 100%;
  border-collapse: separate; 
  border-spacing: 0 1rem;
  font-size: 0.875rem;
  margin-top: 0;
  .param-list {
    margin-top: 0;
  }

  tr {
    border: none !important;
  }

  td {
    padding: 0.5rem 1rem;
    border: none !important;
    margin: 0.5rem 0;
    vertical-align: middle;
    &:first-child {
      border-left: 3px solid #f0f0f0 !important;
      vertical-align: top;
    }
  }

  thead {
    border: none !important;
  }

  th {
    border: none !important;
  }
}

.param-list {
  td:first-child {
    padding-left: 1rem !important;
    min-width: 6rem;
  }
}

@media (max-width: $MQMobile) {
  .param-list {
    td {
      display: block;
      border-left: 3px solid #f0f0f0 !important;
      margin: 0;
      padding-bottom: 0;
      color: #666;
      &:first-child {
        color: $textColor;
      }
      &:last-child {
        padding-bottom: 0.5rem;
      }
    }
  }
}
</style>

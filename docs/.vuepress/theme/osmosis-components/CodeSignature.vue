<template>
  <div class="code">
    <code>
      {{ path }}.
      <b>{{ signature }}</b>
    </code>(delegator=None, validator=None)
  </div>
</template>

<script>
export default {
  name: "CodeSignature",
  props: {
    path: String,
    method: String
  },
  data: function() {
    return {
      signature: this.path
    };
  }
};
</script>

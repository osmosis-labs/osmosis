<template>
  <div>
    <button @click="testAlert">{{ displayText }}</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      dynamicComponent: null
    };
  }
};
</script>

<style scoped>
button {
  background-color: #4caf50;
}
</style>
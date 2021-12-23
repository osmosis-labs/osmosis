<template>
  <div class="errorPage">
    <h1>Oops!</h1>

    <blockquote>{{ getMsg() }}</blockquote>
    <div>
      <img src="/img/404page.png" />
    </div>

    <RouterLink to="/">
      <div class="button">
        Back to Osmosis Docs
        <i class="material-icons">redo</i>
      </div>
    </RouterLink>
  </div>
</template>

<script>
const msgs = [
  `There's nothing here.`,
  `How did we get here?`,
  `Did we mix the wrong pool?`,
  `Wrong formula!.`
];

export default {
  methods: {
    getMsg() {
      return msgs[Math.floor(Math.random() * msgs.length)];
    }
  }
};
</script>

<style lang="stylus">
.errorPage {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 2rem;
  overflow-y: auto;

  h1 {
    padding: 0;
    margin: 0;
    font-size: 3rem;
  }

  img {
    max-width: 400px;
    width: 65vw;
    height: auto;
  }

  .button {
    margin-top: 1rem;
    background-color: $primaryColor;
    color: #fff;
    border-radius: 0.625rem;
    padding: 1rem 1.875rem;
    display: flex;
    justify-content: center;
    align-items: center;
    font-weight: 500;

    i {
      margin-left: 0.5rem;
    }
  }
}

@media (max-width: $MQMobile) {
  .errorPage {
    justify-content: flex-start;

    img {
      max-width: 300px;
    }
  }
}
</style>

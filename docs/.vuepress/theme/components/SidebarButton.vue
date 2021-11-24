<template>
  <div class="sidebar-button" @click="$emit('toggle-sidebar')">
  </div>
</template>

<style lang="stylus">
.sidebar-open {
  .sidebar-button {
    &:before {
      content 'close'
    }
  }
}
.sidebar-button {
  cursor: pointer;
  display: none;
  width: 2.25rem
  height: 2.25rem;
  border-radius: 1.125rem;
  background: rgba(32,67,181,0.08)
  position: absolute;
  padding: 0;
  text-align: center;
  box-sizing: border-box;
  top: $mobile-navbar-vertical-padding;
  right: 2rem;
  color: $primaryColor;

  &:before {
    content 'menu'
    display block
    color $primaryColor
    font-family 'Material Icons'
    font-weight normal
    font-style normal
    font-size 1.25rem
    text-align center
    line-height 2.25rem
    text-transform none
    letter-spacing normal
    word-wrap normal
    white-space nowrap
    direction ltr
    -webkit-font-smoothing antialiased
    text-rendering optimizeLegibility
    -moz-osx-font-smoothing grayscale
    font-feature-settings 'liga'
  }
}

@media (max-width: $MQMobile) {
  .sidebar-button {
    display: block;
  }
}

@media (max-width: $MQMobileNarrow) {
  .sidebar-button {
    display: block;
    right: 1.5rem;
  }
}
</style>

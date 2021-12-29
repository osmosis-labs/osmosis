<template>
  <main class="page">
    <slot name="top" />
    <Breadcrumb />
    <Content class="theme-default-content" />
    <PageEdit />

    <PageNav v-bind="{ sidebarItems }" />

    <slot name="bottom" />
    <Footer />
  </main>
</template>

<script>
import PageEdit from "@theme/components/PageEdit.vue";
import PageNav from "@theme/components/PageNav.vue";
import Footer from "@theme/components/Footer.vue";
import Breadcrumb from "@theme/components/Breadcrumb.vue";

export default {
  components: { PageEdit, PageNav, Footer, Breadcrumb },
  props: ["sidebarItems"]
};
</script>

<style lang="stylus">
@require '../styles/wrapper.styl';

.page {
  padding-bottom: 2rem;
  display: block;
}
</style>

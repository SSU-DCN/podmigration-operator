<template lang="pug">
  div
    v-app-bar(:clipped-left='$vuetify.breakpoint.mdAndUp' color='primary' app)
      v-app-bar-nav-icon(@click.stop='drawer = !drawer')
      span.title.ml-3.mr-5
        | Podmigration
      v-text-field.hidden-sm-and-down(
        v-if='showSearch'
        prepend-inner-icon='mdi-magnify'
        label='Search sources...'
        solo-inverted
        hide-details
        flat
      )
      v-spacer
      v-btn(icon)
        v-icon mdi-bell
      v-btn(icon large)
        v-avatar(size='32px' tile)
          img(src='https://cdn.vuetifyjs.com/images/logos/logo.svg', alt='avatar')
    v-navigation-drawer(v-model='drawer' :clipped='$vuetify.breakpoint.mdAndUp' app)
      v-list
        v-list-item(v-for='(item, i) in items' :key="i" :to='item.path')
          v-list-item-icon
            v-icon {{ item.icon }}
          v-list-item-content
            v-list-item-title {{ item.text }}
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "SidebarHeader",
  data: () => ({
    drawer: null
  }),
  computed: {
    ...mapGetters("app", ["items"]),
    showSearch() {
      return !!this.items.find(
        item => item.search && item.path === this.$route.path
      );
    }
  }
};
</script>

<style scoped></style>

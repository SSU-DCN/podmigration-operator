<template lang="pug">
  v-app
    SidebarHeader
    v-content
      v-container
        router-view
</template>

<script>
import SidebarHeader from "./components/SidebarHeader";

export default {
  name: "App",
  components: { SidebarHeader }
};
</script>

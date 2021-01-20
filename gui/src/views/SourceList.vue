<template lang="pug">
  v-row
    v-col(cols="12")
      v-row.px-6(justify="space-between")
        h4.display-1 All Sources
        v-btn.white--text(color="secondary" :to="{ name: 'add-source' }")
          v-icon(left dark) mdi-plus
          | New Source
    v-row.px-6
      v-col.pa-2(
        v-for="(podmigration, i) in podmigrations"
        :key="`source-${i}`"
        xs="12"
        sm="6"
        xl="4"
      )
        PodmigrationCard(:podmigration="podmigration")
      v-col.text-sm-center(v-if="!podmigrations.length")
        h5.headline.grey--text No sources yet!
</template>

<script>
import PodmigrationCard from "../components/PodmigrationCard";

export default {
  name: "SourceList",
  title() {
    return this.$options.name;
  },
  components: { PodmigrationCard },
  data: () => ({
    podmigrations: []
  }),
  created() {
    this.$api
      .getAllPodmigrations()
      .then(list => (this.podmigrations = list.items))
      .catch(e => console.log(e));
  }
};
</script>

<style scoped></style>

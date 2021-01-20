<template lang="pug">
  div
    v-toolbar(:color='statusColor("deployed")' dark)
      v-btn.hidden-xs-only(icon)
        v-icon mdi-image-search-outline
      v-toolbar-title {{ podmigration.name }}
      v-spacer
      v-btn(:to="{ name: 'source-detail', params: { name: podmigration.name } }" small outlined)
        | View
        v-icon(right='') mdi-share
    v-card
      v-card-text
        p Domains:
        ul
          li(v-for='(domain, i) in podmigration.status.domains' :key='i') {{ domain }}
      v-card-actions.pa-3
        v-icon(:color='statusColor("deployed")' small) mdi-checkbox-blank-circle
        span.ml-2 {{ "deployed" | capitalize }}
        v-spacer
          | {{ podmigration.type }}
          v-icon.ml-2(small='') {{ iconClass(podmigration.type) }}
</template>

<script>
import statusColorMap from "./utils/statusColorMap";
import iconClassMap from "./utils/iconClassMap";

export default {
  name: "PodmigrationCard",
  props: {
    podmigration: { type: Object, required: true }
  },
  methods: {
    statusColor: val => statusColorMap.get(val),
    iconClass: val => iconClassMap.get(val)
  }
};
</script>

<style scoped></style>

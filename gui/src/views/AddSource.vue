<template lang="pug">
  div
    v-row
      v-col(cols="12")
        h1.display-1.mb-4.font-weight-medium New Source
        v-divider
      v-col
        v-row
          v-col(cols="12")
            h3.headline General
            v-row
              v-col(cols="12" md="6")
                v-text-field(
                  v-model="name"
                  label="Source Name"
                  outlined
                )
              v-col(cols="12" md="6")
                v-select(
                  v-model="action"
                  label="Action"
                  :items="SOURCE_ITEMS"
                  :prepend-icon="iconClass(action)"
                  outlined
                )
            v-divider.mb-6
            h3.headline {{ SOURCE_ITEMS.find(el => el.value === action).text }} Settings
            v-row(v-if="action === 'restore'")
              v-col(cols="12" md="6")
                v-text-field(
                  v-model="replicas"
                  label="Replicas"
                  hide-details="auto"
                  dense
                  outlined
                )
              v-col(cols="12" md="8")
                v-text-field(
                  v-model="snapshotPath"
                  label="Snapshot Path"
                  placeholder="/var/lib/kubelet/migration/pod-name"
                  hide-details="auto"
                  outlined
                )
            v-row(v-if="action === 'checkpoint'")
              v-col(cols="12" md="6")
                v-text-field(
                  v-model="sourcePod"
                  label="Source Pod"
                  hide-details="auto"
                  dense
                  outlined
                )
              v-col(cols="12" md="8")
                v-text-field(
                  v-model="snapshotPath"
                  label="Snapshot Path"
                  placeholder="/var/lib/kubelet/migration/"
                  hide-details="auto"
                  outlined
                )
            v-row(v-else-if="action === 'live-migration'")
              v-col(cols="12" md="6")
                v-text-field(
                  v-model="sourcePod"
                  label="Source Pod"
                  hide-details="auto"
                  dense
                  outlined
                )
              v-col(cols="12" md="8")
                v-text-field(
                  v-model="destHost"
                  label="Target Node"
                  hide-details="auto"
                  dense
                  outlined
                )
            v-divider.my-6
            h3.headline Namespace
            v-row
              v-col(cols="12" md="6")
                v-text-field(
                  v-model="Namespace"
                  placeholder="default"
                  hide-details="auto"
                  dense
                  outlined
                  hint="Add subdomain and press enter"
                  append-icon=""
                )
            //- v-divider.mb-6
            //- h3.headline Cache Settings
            //- v-row
            //-   v-col(cols="12" md="6")
            //-     v-select(
            //-       v-model="cacheTTL"
            //-       :items="CACHE_TTL_BEHAVIOR"
            //-       label="Cache TTL Behavior"
            //-       outlined
            //-       hide-details="auto"
            //-     )
            //-   v-col(cols="12" md="6")
            //-     v-text-field(
            //-       v-model="defaultCache"
            //-       type="number"
            //-       label="Default Cache TTL (seconds)"
            //-       outlined
            //-       hide-details="auto"
            //-     )
        v-row(justify="end")
          v-col(cols="auto")
            v-btn.ma-2(large color="error" :to="{ name: 'sources' }")
              v-icon(left) mdi-cancel
              | Cancel
            v-btn.ma-2(large color="primary" @click="doCreate")
              v-icon(left) mdi-content-save-outline
              | Save
</template>

<script>
import iconClassMap from "../components/utils/iconClassMap";

export default {
  name: "AddSource",
  title: "Add New Source",
  data: () => ({
    name: "",
    sourcePod: "",
    action: "checkpoint",
    snapshotPath: "",
    replicas: 1,
    destHost: "",
    Namespace: "default"
  }),
  computed: {
    SOURCE_ITEMS: () => [
      { text: "live-migration", value: "live-migration" },
      { text: "checkpoint", value: "checkpoint" },
      { text: "restore", value: "restore" }
    ],
    CACHE_TTL_BEHAVIOR: () => [
      { text: "Respect Origin", value: "RespectOrigin" },
      { text: "Override Origin", value: "OverrideOrigin" },
      { text: "Enforce Minimum", value: "EnforceMinimum" }
    ]
  },
  methods: {
    iconClass: val => iconClassMap.get(val),
    doCreate() {
      this.$api.createPodmigration(this.payload()).then(() => {
        this.$router.replace({ name: "sources" });
      });
    },
    payload() {
      if (this.action == "live-migration") {
        return {
          name: this.name,
          replicas: 1,
          action: "live-migration",
          sourcePod: this.sourcePod,
          destHost: this.destHost,
          Namespace: this.Namespace
        };
      } else if (this.action == "restore") {
        return {
          name: this.name,
          action: "restore",
          replicas: parseInt(this.replicas),
          snapshotPath: this.snapshotPath,
          Namespace: this.Namespace
        };
      } else if (this.action == "checkpoint") {
        return {
          name: this.name,
          replicas: 1,
          action: "checkpoint",
          sourcePod: this.sourcePod,
          snapshotPath: this.snapshotPath,
          Namespace: this.Namespace
        };
      }
    }
  }
};
</script>

<style scoped></style>

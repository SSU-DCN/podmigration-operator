import Vue from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import "./mixins";
import "./filters";
import vuetify from "./plugins/vuetify";
import "roboto-fontface/css/roboto/roboto-fontface.css";
import "@mdi/font/css/materialdesignicons.css";
import Podmigration from "./services/podmigration";

Vue.config.productionTip = false;

function VUE_APP() {
  // setup the HTTP API namespace
  Vue.prototype.$api = new Podmigration();

  new Vue({
    router,
    store,
    vuetify,
    render: h => h(App)
  }).$mount("#app");
}

function SETUP() {
  VUE_APP();
}

SETUP();

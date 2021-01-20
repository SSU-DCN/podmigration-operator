import Vue from "vue";
import VueRouter from "vue-router";
import SourceList from "../views/SourceList";
import SourceDetail from "../views/SourceDetail";
import Dashboard from "../views/Dashboard";
import AddSource from "../views/AddSource";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "dashboard",
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: "/sources",
    name: "sources",
    component: SourceList,
    meta: { requiresAuth: true }
  },
  {
    path: "/sources/new",
    name: "add-source",
    component: AddSource,
    meta: { requiresAuth: true }
  },
  {
    path: "/sources/:name",
    name: "source-detail",
    component: SourceDetail,
    meta: { requiresAuth: true }
  },
  {
    path: "/support",
    name: "support",
    component: () => import("../views/Support")
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

export default router;

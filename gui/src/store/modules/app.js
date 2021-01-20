const state = {
  items: [
    {
      icon: "mdi-desktop-mac-dashboard",
      text: "Dashboard",
      requiresAuth: true,
      path: "/"
    },
    {
      icon: "mdi-image-search",
      text: "Sources",
      search: true,
      requiresAuth: true,
      path: "/sources"
    },
    { icon: "mdi-lifebuoy", text: "Support", path: "/support" }
  ]
};

const getters = {
  items(state) {
    return state.items;
    //   .filter(
    //   e =>
    //     e.requiresAuth === rootGetters["auth/isLoggedIn"] ||
    //     e.requiresAuth == null
    // );
  }
};

const mutations = {};

const actions = {};

export { state, mutations, actions, getters };

const state = {
  isLoggedIn: false,
  thumbnailUrl: null
};

const getters = {
  isLoggedIn: state => state.isLoggedIn,
  thumbnailUrl: state => state.thumbnailUrl
};

const mutations = {
  LOGIN(state) {
    state.isLoggedIn = true;
  },
  LOGOUT(state) {
    state.isLoggedIn = false;
  },
  SET_THUMBNAIL_URL(state, URL) {
    state.thumbnailUrl = URL;
  }
};

const actions = {
  login({ commit }) {
    commit("LOGIN");
  },
  logout({ commit }) {
    commit("LOGOUT");
  }
};

export { state, mutations, actions, getters };

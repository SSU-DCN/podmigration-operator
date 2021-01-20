module.exports = {
  transpileDependencies: ["vuetify"],
  devServer: {
    proxy: {
      "^/api": {
        target: "http://localhost:5000",
        secure: false,
        pathRewrite: { "^/api": "" },
        changeOrigin: true,
        logLevel: "debug"
      }
    }
  }
};

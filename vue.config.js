const HtmlWebpackPlugin = require('html-webpack-plugin');
// const WebpackCdnPlugin = require('webpack-cdn-plugin');

module.exports = {
  lintOnSave: false,
  devServer: {
    proxy: {
      '/ws': {
        target: 'http://localhost:5000/',
        ws: true,
        changeOrigin: false,
      },
      '/api/*': {
        target: 'http://localhost:5000',
      }
    }
  },
  configureWebpack: {
    plugins: [
      new HtmlWebpackPlugin({
        title: "GO Judger",
        template: 'public/index.ejs',
        favicon: 'public/favicon.ico'
      }),
    ]
  }
};

const HtmlWebpackPlugin = require('html-webpack-plugin');
// const WebpackCdnPlugin = require('webpack-cdn-plugin');
const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = {
  lintOnSave: false,
  devServer: {
    proxy: {
      '/api/*': {
        target: 'http://localhost:5000',
      }
    },
    after(app) {
      const wsProxy = createProxyMiddleware({
        target: 'http://localhost:5000/',
        ws: true,
        changeOrigin: false,
      });
      app.use('/api/ws/judge', wsProxy);
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

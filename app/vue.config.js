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
  }
}

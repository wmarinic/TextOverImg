const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  pages: {
    'index': {
      entry: './src/Home/main.js',
      template: 'public/index.html',
      title: 'Home',
      chunks: ['chunk-vendors', 'chunk-common', 'index']
    },
    'register': {
      entry: './src/Register/main.js',
      template: 'public/index.html',
      title: 'Register',
      chunks: ['chunk-vendors', 'chunk-common', 'register']
    }
  }
})



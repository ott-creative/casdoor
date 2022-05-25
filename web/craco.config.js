const CracoLessPlugin = require('craco-less');

module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://sso-api.codegene.xyz',
        changeOrigin: true,
      },
      '/swagger': {
        target: 'http://sso-api.codegene.xyz',
        changeOrigin: true,
      },
      '/files': {
        target: 'http://sso-api.codegene.xyz',
        changeOrigin: true,
      },
      '/.well-known/openid-configuration': {
        target: 'http://sso-api.codegene.xyz',
        changeOrigin: true,
      },
      '/cas/serviceValidate': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/cas/proxyValidate': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/cas/proxy': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/cas/validate': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      }
    },
  },
  plugins: [
    {
      plugin: CracoLessPlugin,
      options: {
        lessLoaderOptions: {
          lessOptions: {
            modifyVars: {'@primary-color': 'rgb(45,120,213)'},
            javascriptEnabled: true,
          },
        },
      },
    },
  ],
};

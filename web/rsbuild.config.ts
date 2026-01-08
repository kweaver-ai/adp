import { defineConfig } from '@rsbuild/core';
import { pluginLess } from '@rsbuild/plugin-less';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginSvgr } from '@rsbuild/plugin-svgr';
import packageJson from './package.json';

const isDev = process.env.NODE_ENV === 'development';

export default defineConfig({
  html: {
    template: './public/index.html',
  },
  dev: {
    assetPrefix: '/vega/',
    client: { protocol: 'ws', host: 'localhost', port: 3010 },
    progressBar: true,
  },
  server: {
    port: 3010,
    open: process.env.FIRST_RUN === '1',
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Cache-Control': 'no-cache',
    },
    proxy: {
      '/api': {
        secure: false,
        changeOrigin: true,
        target: 'https://dip.aishu.cn',
        timeout: 30000,
        proxyTimeout: 30000,
      },
    },
  },
  output: {
    assetPrefix: '/vega/',
    cssModules: {
      auto: true,
      localIdentName: `[local]__[hash:base64:5]`,
    },
    sourceMap: {
      js: isDev ? 'cheap-module-source-map' : false,
      css: isDev,
    },
    cleanDistPath: true,
    dataUriLimit: 10240,
    filename: {
      js: isDev ? '[name].js' : '[name].[hash:8].js',
      css: isDev ? '[name].css' : '[name].[hash:8].css',
      font: 'fonts/[name].[hash:8][ext]',
      image: 'images/[name].[hash:8][ext]',
    },
  },
  plugins: [
    pluginLess(),
    pluginReact(),
    pluginSvgr({
      svgrOptions: {
        // 将 SVG 作为图标处理（自动设置 width/height 为 1em 等）
        icon: true,
      },
    }),
  ],
  performance: {
    removeConsole: !isDev,
    removeMomentLocale: true,
  },
  tools: {
    rspack: {
      output: {
        library: `${packageJson.name}-[name]`,
        libraryTarget: 'umd',
        chunkLoadingGlobal: `webpackJsonp_${packageJson.name}`,
      },
      optimization: {
        minimize: !isDev,
        runtimeChunk: 'single',
        splitChunks: {
          chunks: 'all',
          cacheGroups: {
            react: {
              test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
              name: 'react-vendor',
              priority: 10,
              reuseExistingChunk: true,
            },
            antd: {
              test: /[\\/]node_modules[\\/]antd[\\/]/,
              name: 'antd-vendor',
              priority: 9,
              reuseExistingChunk: true,
            },
            antv_x6: {
              test: /[\\/]node_modules[\\/]@antv[\\/]x6[\\/]/,
              name: 'antv-x6-vendor',
              priority: 8,
              reuseExistingChunk: true,
            },
            antv_g6: {
              test: /[\\/]node_modules[\\/]@antv[\\/]g6[\\/]/,
              name: 'antv-g6-vendor',
              priority: 8,
              reuseExistingChunk: true,
            },
            antv_other: {
              test: /[\\/]node_modules[\\/]@antv[\\/](?!x6|g6)(.*?)[\\/]/,
              name: 'antv-other-vendor',
              priority: 7,
              reuseExistingChunk: true,
            },
            lodash: {
              test: /[\\/]node_modules[\\/]lodash-es[\\/]/,
              name: 'lib-lodash',
              priority: 11,
              reuseExistingChunk: true,
            },
            echarts: {
              test: /[\\/]node_modules[\\/]echarts[\\/]/,
              name: 'lib-echarts',
              priority: 11,
              reuseExistingChunk: true,
            },
            codemirror: {
              test: /[\\/]node_modules[\\/](@codemirror|@uiw[\\/](react-)?codemirror)[\\/]/,
              name: 'lib-codemirror',
              priority: 11,
              reuseExistingChunk: true,
            },
            monaco: {
              test: /[\\/]node_modules[\\/](@monaco-editor|monaco-editor)[\\/]/,
              name: 'lib-monaco',
              priority: 11,
              reuseExistingChunk: true,
            },
            jsoneditor: {
              test: /[\\/]node_modules[\\/]jsoneditor[\\/]/,
              name: 'lib-jsoneditor',
              priority: 11,
              reuseExistingChunk: true,
            },
            aishu: {
              test: /[\\/]node_modules[\\/]@aishu-tech[\\/]/,
              name: 'lib-aishu',
              priority: 11,
              reuseExistingChunk: true,
            },
            xyflow: {
              test: /[\\/]node_modules[\\/]@xyflow[\\/]/,
              name: 'lib-xyflow',
              priority: 11,
              reuseExistingChunk: true,
            },
            immutable: {
              test: /[\\/]node_modules[\\/]immutable[\\/]/,
              name: 'lib-immutable',
              priority: 11,
              reuseExistingChunk: true,
            },
            common: {
              minChunks: 3,
              minSize: 30000,
              maxSize: 500000,
              priority: 5,
              reuseExistingChunk: true,
              name: 'common',
            },
          },
        },
      },
    },
  },
});

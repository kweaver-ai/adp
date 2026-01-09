import ReactDOM from 'react-dom/client';
import 'react-resizable/css/styles.css';
import api from '@/services/authorization';
import { baseConfig } from '@/services/request';
import App from '@/pages/router';
import '@/public-path';
import '@/style/reset.less';
// eslint-disable-next-line import/order
import '@/style/global.less';
import UTILS from '@/utils';

// 过滤掉 findDOMNode 警告
const originalError = console.error;
console.error = (...args: any[]) => {
  if (typeof args[0] === 'string' && args[0].includes('findDOMNode')) {
    return;
  }
  originalError.call(console, ...args);
};

const init = async () => {
  try {
    const userPermissionOperation = await api.getResourceTypeOperation();
    sessionStorage.setItem('vega.userPermissionOperation', JSON.stringify(userPermissionOperation || []));
  } catch (error) {
    console.log('error: 获取权限失败', error);
  }
};

let root: any;
const render = (props: any) => {
  const container = document.getElementById('vega-root');
  if (!container) return;

  root = root || ReactDOM.createRoot(container);
  root.render(<App {...props} />);
};

if (!(window as any).__POWERED_BY_QIANKUN__) {
  // 本地调试，手动填入token
  baseConfig.lang = 'zh-cn';
  baseConfig.token = 'ory_at_Qr06Od4elSeVIQj9CBVM7JnC83qk2oVMjp9bCw6WpyY.I3VkF2kyBy7Of-wOGGbhIhC9kXVhne0XUY00LvDVpNc';
  baseConfig.userid = '488c973e-6f67-11f0-b0dc-36fa540cff80';
  baseConfig.roles = [];
  baseConfig.toggleSideBarShow = () => {};
  baseConfig.businessDomainID = 'bd_public';
  UTILS.SessionStorage.set('language', 'zh-cn');
  UTILS.SessionStorage.set('token', baseConfig.token);

  init().finally(() => {
    render({});
  });
}

export async function bootstrap(props: any) {
  baseConfig.lang = props?.lang;
  baseConfig.token = props?.token?.getToken?.access_token;
  baseConfig.userid = props?.userid;
  baseConfig.roles = props?.config?.userInfo?.user?.roles || [];
  baseConfig.refresh = props?.token?.refreshOauth2Token;
  baseConfig.toggleSideBarShow = props?.toggleSideBarShow;
  baseConfig.businessDomainID = props?.businessDomainID || '';
  baseConfig.history = props?.history;
  baseConfig.navigate = props?.navigate;
  UTILS.SessionStorage.set('language', props?.lang);
  UTILS.SessionStorage.set('token', props?.token?.getToken?.access_token);
}
export async function mount(props: any) {
  await Promise.resolve();
  init().finally(() => {
    render(props);
  });
}
export async function unmount() {
  if (root) {
    root.unmount();
    root = null;
  }
}

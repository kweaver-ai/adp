import axios, { AxiosRequestConfig, Method, AxiosResponse } from 'axios';
import UTILS from '@/utils';

// 定义配置接口
export interface BaseConfig {
  lang: string;
  token: string;
  userid: string;
  roles: { id: string; name: string }[];
  refresh: () => Promise<any>;
  toggleSideBarShow: (bool?: boolean) => void;
  businessDomainID: string;
  history: any;
  navigate: any;
}

export const baseConfig: BaseConfig = {
  lang: 'zh-cn',
  token: '',
  userid: '',
  roles: [],
  refresh: () => Promise.resolve(),
  toggleSideBarShow: () => {},
  businessDomainID: '',
  history: null,
  navigate: null,
};

// 取消请求 token 存储
const cancelSources: Record<string, any> = {};

// 创建 axios 实例
const service = axios.create({ baseURL: '/', timeout: 20000 });

// 请求拦截器
service.interceptors.request.use(
  (config: any) => {
    // 1. 添加取消请求机制
    if (config.cancelTokenKey) {
      config.cancelToken = new axios.CancelToken((cancel) => {
        cancelSources[config.cancelTokenKey] = cancel;
      });
    }

    // 2. 设置基础 Headers
    config.headers['Content-Type'] = 'application/json; charset=utf-8';
    config.headers['Accept-Language'] = baseConfig.lang === 'en-us' ? 'en-us' : 'zh-cn';

    // 3. 注入 Token
    if (baseConfig.token) {
      config.headers.Authorization = `Bearer ${baseConfig.token}`;
    } else if (!config.headers.Authorization) {
      const token = UTILS.SessionStorage.get('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    // 4. 处理文件上传 (type: 'file')
    if (config?.data?.type === 'file') {
      config.headers['Content-Type'] = 'multipart/form-data';
      const formData = new FormData();

      const { type: _type, ...otherData } = config.data;

      Object.entries(otherData).forEach(([key, value]) => {
        if (key === 'file' && Array.isArray(value)) {
          value.forEach((fileItem) => formData.append('file', fileItem));
        } else {
          formData.append(key, value as any);
        }
      });
      config.data = formData;
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    // 清理 cancelToken
    const key = (response?.config as any)?.cancelTokenKey;
    if (key && cancelSources[key]) {
      delete cancelSources[key];
    }
    return response;
  },
  (error) => {
    if (axios.isCancel(error)) {
      return { Code: -200, message: '取消请求', cause: '取消请求' };
    }
    return Promise.reject(error);
  }
);

// 请求参数接口
interface CustomConfig extends AxiosRequestConfig {
  isNoHint?: boolean;
  cancelTokenKey?: string;
}

interface RequestParams {
  url: string;
  method: Method;
  data?: any;
  config?: CustomConfig;
  reTry?: boolean;
}

/**
 * 通用请求函数
 */
const request = <T>({ url, method, data = {}, config = { isNoHint: false }, reTry = false }: RequestParams): Promise<T> => {
  // GET 请求参数放在 params 中，其他请求放在 data 中
  const requestBody = method.toUpperCase() === 'GET' ? { params: data } : { data };

  return new Promise<T>((resolve, reject) => {
    const { isNoHint, ...restConfig } = config;

    service({ url, method, ...requestBody, ...restConfig })
      .then((response: AxiosResponse) => {
        if (response) {
          resolve(response.data);
        } else {
          resolve(response as unknown as T); // 处理 cancel 等特殊情况返回非标准 response
        }
      })
      .catch((error) => {
        const { data: responseData, status } = error.response || {};
        const { description } = responseData || {};

        // 401 处理：尝试刷新 Token 并重试
        if (status === 401) {
          if (reTry) {
            UTILS.message.error(description || 'Unauthorized');
            return reject(error.response);
          }

          baseConfig
            .refresh()
            .then((result) => {
              if (!result || !result.access_token) {
                UTILS.message.error(description || 'token无效或已过期');
                return reject(error);
              }
              // 更新 Token
              baseConfig.token = result.access_token;
              UTILS.SessionStorage.set('token', result.access_token);

              // 重试请求
              request<T>({ url, method, data, config, reTry: true }).then(resolve).catch(reject);
            })
            .catch((refreshErr) => {
              UTILS.message.error(description || 'Session expired');
              reject(refreshErr.response || error.response);
            });
        }
        // 静默模式：直接返回错误数据
        else if (isNoHint) {
          resolve(responseData);
        }
        // 默认错误处理：弹窗提示
        else {
          const errorMessage = description || 'Error';
          // 避免在 cancel 时弹出错误
          if (!axios.isCancel(error)) {
            UTILS.message.error(errorMessage);
          }
          reject(error.response);
        }
      });
  });
};

// 封装快捷方法
const requestGet = <T>(url: string, data?: any, config: CustomConfig = {}): Promise<T> => request<T>({ url, data, config, method: 'GET' });
// POST/PUT/DELETE: data 参数为 body (如果需要 query params，请在 config.params 中传递)
const requestPost = <T>(url: string, data?: any, config: CustomConfig = {}): Promise<T> => request<T>({ url, data, config, method: 'POST' });
const requestPut = <T>(url: string, data?: any, config: CustomConfig = {}): Promise<T> => request<T>({ url, data, config, method: 'PUT' });
const requestDel = <T>(url: string, data?: any, config: CustomConfig = {}): Promise<T> => request<T>({ url, data, config, method: 'DELETE' });
const requestPostOverrideGet = <T>(url: string, data?: any, config: CustomConfig = {}): Promise<T> => {
  return request<T>({
    url,
    data,
    config: { ...config, headers: { ...config.headers, 'x-http-method-override': 'GET' } },
    method: 'POST',
  });
};

const Request = {
  get: requestGet,
  post: requestPost,
  put: requestPut,
  delete: requestDel,
  postOverrideGet: requestPostOverrideGet,
  cancels: cancelSources,
};

export default Request;

import { useEffect, lazy, Suspense } from 'react';
import intl from 'react-intl-universal';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import '@aishu-tech/components/dist/dip-components.full.css';
import { message, ConfigProvider, ThemeConfig, Spin } from 'antd';
import enUS from 'antd/lib/locale/en_US';
import zhCN from 'antd/lib/locale/zh_CN';
import HOOKS from '@/hooks';
import locales from '@/locales';
import THEME from '@/theme.ts';
import UTILS from '@/utils';
import { Modal } from '@/web-library/common';

// 异步加载 @aishu-tech/components
let aishuComponentsPromise: Promise<any> | null = null;
const loadAishuComponents = () => {
  if (!aishuComponentsPromise) {
    aishuComponentsPromise = import('@aishu-tech/components/dist/dip-components.min.js');
  }
  return aishuComponentsPromise;
};

const ActionCreateAndEdit = lazy(() => import('./ActionCreateAndEdit'));
const ActionDetail = lazy(() => import('./Action/Detail'));
const AtomDataView = lazy(() => import('./AtomDataView'));
const CustomDataView = lazy(() => import('./CustomDataView'));
const CustomDataViewDetailContent = lazy(() => import('./CustomDataView/MainContent/DetailContent'));
const DataConnectForm = lazy(() => import('./DataConnect/DataConnectForm'));
const DataConnect = lazy(() => import('./DataConnect/tabs'));
const EdgeCreateAndEdit = lazy(() => import('./EdgeCreateAndEdit'));
const KnowledgeNetwork = lazy(() => import('./KnowledgeNetwork'));
const KnowledgeNetworkMain = lazy(() => import('./KnowledgeNetworkMain'));
const MetricModel = lazy(() => import('./MetricModel'));
const FormContainer = lazy(() => import('./MetricModel/FormContainer'));
const ObjectCreateAndEdit = lazy(() => import('./ObjectCreateAndEdit'));
const ObjectIndexSetting = lazy(() => import('./ObjectIndexSetting'));
const RowColumnPermission = lazy(() => import('./RowColumnPermission'));

interface AppProps {
  protocol?: string;
  host?: string;
  port?: number | string;
  lang?: string;
  container?: HTMLElement;
  token?: any;
  prefix?: string;
  oemConfigs?: any;
  history?: any;
  [key: string]: any;
}

const TITLE: Record<string, string> = { 'zh-cn': '工作站', 'en-us': 'Studio', 'zh-tw': '工作站' };
const App = (props: AppProps) => {
  const [modal, modalContextHolder] = Modal.useModal();
  const [messageApi, messageContextHolder] = message.useMessage();
  const { lang: language = 'zh-cn', container, token, prefix = '', oemConfigs } = props;
  const { protocol = 'https:', hostname, port = 443 } = props?.config?.systemInfo?.location || {};

  useEffect(() => {
    document.title = TITLE[language];
    message.config({
      top: 32,
      maxCount: 1,
      getContainer: () => document.getElementById('vega-root') || container!,
    });
    intl.init({ currentLocale: language, locales, warningHandler: () => '' });
    UTILS.initMessage(messageApi);

    // 异步加载并初始化 @aishu-tech/components
    loadAishuComponents().then((module) => {
      const { apis } = module;
      apis.setup({
        protocol,
        host: hostname,
        port,
        lang: language,
        getToken: () => token?.getToken.access_token,
        prefix,
        theme: oemConfigs?.theme,
        popupContainer: document.getElementById('vega-root'),
        refreshToken: token?.refreshOauth2Token,
        onTokenExpired: token?.onTokenExpired,
      });
    });
  }, []);

  return (
    <ConfigProvider
      locale={language === 'en-us' ? enUS : zhCN}
      wave={{ disabled: true }}
      theme={THEME as ThemeConfig}
      getPopupContainer={() => document.getElementById('vega-root') || container!}
      getTargetContainer={() => document.getElementById('vega-root') || container!}
    >
      <HOOKS.GlobalProvider value={{ modal, message: messageApi, baseProps: props || {} }}>
        {modalContextHolder}
        {messageContextHolder}
        <Router basename={(window as any).__POWERED_BY_QIANKUN__ ? props.history.getBasePath : '/vega'}>
          <Suspense
            fallback={
              <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <Spin />
              </div>
            }
          >
            <Switch>
              <Route exact path="/" render={() => <KnowledgeNetwork />} />
              <Route exact path="/ontology" render={() => <KnowledgeNetwork />} />
              <Route path="/ontology/main" render={() => <KnowledgeNetworkMain />} />
              <Route exact path="/ontology/edge/create" render={() => <EdgeCreateAndEdit />} />
              <Route exact path="/ontology/edge/edit/:id" render={() => <EdgeCreateAndEdit />} />
              <Route exact path="/ontology/object/create" render={() => <ObjectCreateAndEdit />} />
              <Route exact path="/ontology/object/edit/:id" render={() => <ObjectCreateAndEdit />} />
              <Route exact path="/ontology/action/create" render={() => <ActionCreateAndEdit />} />
              <Route exact path="/ontology/action/edit/:id" render={() => <ActionCreateAndEdit />} />
              <Route path="/ontology/action/detail/:knId/:atId" render={() => <ActionDetail />} />
              <Route exact path="/ontology/object/settting/:id" render={() => <ObjectIndexSetting />} />
              <Route exact path="/metric-model" render={() => <MetricModel />} />
              <Route exact path="/metric-model/create/:createType" render={() => <FormContainer />} />
              <Route exact path="/metric-model/edit/:id" render={() => <FormContainer />} />
              <Route exact path="/custom-data-view" render={() => <CustomDataView />} />
              <Route exact path="/custom-data-view/detail/:id?" render={() => <CustomDataViewDetailContent />} />
              <Route exact path="/atom-data-view" render={() => <AtomDataView />} />
              <Route exact path="/data-connect" render={() => <DataConnect />} />
              <Route exact path="/data-connect/create" render={() => <DataConnectForm />} />
              <Route exact path="/data-connect/edit/:id" render={() => <DataConnectForm />} />
              <Route exact path="/custom-data-view/row-column-permission/:id" render={() => <RowColumnPermission />} />
              <Route render={() => <div>not found</div>} />
            </Switch>
          </Suspense>
        </Router>
      </HOOKS.GlobalProvider>
    </ConfigProvider>
  );
};

export default App;

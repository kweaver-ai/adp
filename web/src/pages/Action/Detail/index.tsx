import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Route, Switch, useHistory, useLocation, useParams, useRouteMatch, Redirect } from 'react-router-dom';
import { LeftOutlined } from '@ant-design/icons';
import { baseConfig } from '@/services/request';
import { Button, message } from 'antd';
import classnames from 'classnames';
import actionApi from '@/services/action';
import * as ActionType from '@/services/action/type';
import { IconFont } from '@/web-library/common';
import styles from './index.module.less';
import Overview from './Overview';
import TaskManagement from './TaskManagement';

interface TLinkItem {
  icon: string;
  title: string;
  url: string;
  count?: number;
  isActive: boolean;
  style?: React.CSSProperties;
}

const LinkItem = (props: TLinkItem) => {
  const history = useHistory();
  const { icon, title, count, url, isActive, style } = props;
  const isShowCount = count !== undefined && count !== null;
  const curCount = count && count > 9999 ? '9999+' : count;

  const toPath = () => {
    history.push(url);
  };

  return (
    <div className={classnames(styles['side-item'], isActive && styles['side-item-active'])} onClick={toPath} style={style}>
      <dl>
        <dt>
          <IconFont type={icon} style={{ color: isActive ? '#126ee3' : '#000', fontSize: 16 }} />
        </dt>
        <dd>{title}</dd>
      </dl>
      {isShowCount && <span className={styles['side-item-count']}>{curCount}</span>}
    </div>
  );
};

const ActionDetail = () => {
  const history = useHistory();
  const { path, url } = useRouteMatch();
  const { pathname } = useLocation();
  const { knId, atId } = useParams<{ knId: string; atId: string }>();
  const [active, setActive] = useState<string>('overview');
  const [detail, setDetail] = useState<any>();
  const [executing, setExecuting] = useState(false);
  const [refreshTask, setRefreshTask] = useState(false);

  const getDetail = async () => {
    try {
      const [res] = await actionApi.getActionTypeDetail(knId, [atId]);
      setDetail(res);
    } catch (error) {
      console.error(error);
    }
  };

  const goback = () => {
    history.push(`/ontology/main/action?id=${knId}`);
  };

  useEffect(() => {
    const last = pathname.split('/').pop() as string;
    setActive(last);
  }, [pathname]);

  useEffect(() => {
    baseConfig?.toggleSideBarShow(false);
    if (knId && atId) {
      getDetail();
    }
  }, [knId, atId]);

  const handleExecute = async () => {
    setExecuting(true);
    try {
      const request: ActionType.ActionExecutionRequest = {
        unique_identities: [],
      };
      await actionApi.executeActionType(knId, atId, request);
      message.success(intl.get('Action.executeSuccess'));
      setRefreshTask(true);
    } catch (error) {
      console.error('Execute error:', error);
      message.error(intl.get('Action.executeFailed'));
    } finally {
      setExecuting(false);
    }
  };

  return (
    <div className={styles['main-box']}>
      <div className={styles['main-header']}>
        <LeftOutlined onClick={goback} />
        <div className={styles['name-icon']}>
          <IconFont type="icon-dip-hangdonglei" style={{ fontSize: 24 }} />
        </div>
        <h4>{detail?.name}</h4>
        <Button type="primary" style={{ marginLeft: 'auto' }} onClick={handleExecute} loading={executing}>
          {intl.get('Action.executeImmediately')}
        </Button>
      </div>
      <div className={styles['main-layout']}>
        <div className={styles['main-side']}>
          <LinkItem icon="icon-dip-tail" isActive={active === 'overview'} title={intl.get('KnowledgeNetwork.overview')} url={`${url}/overview`} />
          <LinkItem icon="icon-dip-task" isActive={active === 'task'} title={intl.get('Global.taskManagement')} url={`${url}/task`} />
        </div>
        <div className={styles['main-content']}>
          <Switch>
            <Route exact path={path}>
              <Redirect to={`${url}/overview`} />
            </Route>
            <Route exact path={`${path}/overview`} render={() => <Overview knId={knId} atId={atId} detail={detail} />} />
            <Route
              exact
              path={`${path}/task`}
              render={() => <TaskManagement knId={knId} atId={atId} refreshTask={refreshTask} onRefreshComplete={() => setRefreshTask(false)} />}
            />
            <Route render={() => <div>{intl.get('Global.pageNotFound')}</div>} />
          </Switch>
        </div>
      </div>
    </div>
  );
};

export default ActionDetail;

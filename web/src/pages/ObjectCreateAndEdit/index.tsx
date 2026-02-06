import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import intl from 'react-intl-universal';
import { useHistory, useParams } from 'react-router-dom';
import { Form } from 'antd';
import HeaderSteps from '@/components/HeaderSteps';
import { queryConceptGroups } from '@/services/conceptGroup';
import * as ConceptGroupType from '@/services/conceptGroup/type';
import * as OntologyObjectType from '@/services/object/type';
import { baseConfig } from '@/services/request';
import HOOKS from '@/hooks';
import SERVICE from '@/services';
import BasicInformation from './BasicInformation';
import DataAttribute from './DataAttribute';
import styles from './index.module.less';
import LogicAttribute from './LogicAttribute';

const ObjectCreateAndEdit = () => {
  const history = useHistory();
  const params: { id: string } = useParams();
  const { message } = HOOKS.useGlobalContext();
  const { id } = params || {};
  const isEditPage = !!id;
  const knId = localStorage?.getItem('KnowledgeNetwork.id') || '';
  const steps = [{ title: intl.get('Global.basicInfo') }, { title: intl.get('Global.dataProperty') }, { title: intl.get('Object.logicProperty') }];

  const [loading, setLoading] = useState(false);
  const [stepsCurrent, setStepsCurrent] = useState(0);
  const [doneStep, setDoneStep] = useState(0);
  const [basicValue, setBasicValue] = useState<OntologyObjectType.BasicInfo>();
  const [dataSource, setDataSource] = useState<OntologyObjectType.DataSource>();
  const [dataProperties, setDataProperties] = useState<OntologyObjectType.DataProperty[]>([]);
  const [logicProperties, setLogicProperties] = useState<OntologyObjectType.LogicProperty[]>([]);
  const [conceptGroups, setConceptGroups] = useState<ConceptGroupType.BasicInfo[]>([]);
  const [conceptGroupsLoading, setConceptGroupsLoading] = useState(false);
  const [basicForm] = Form.useForm();
  const logicAttributeRef = useRef<any>(null);
  const dataAttributeRef = useRef<any>(null);
  const [primaryKeys, setPrimaryKeys] = useState<string[]>([]);
  const [displayKey, setDisplayKey] = useState<string>('');
  const [incrementalKey, setIncrementalKey] = useState<string>('');

  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = intl.get('Global.confirmBackContent');
      return e.returnValue;
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, []);

  useEffect(() => {
    baseConfig?.toggleSideBarShow(false);

    const fetchConceptGroups = async () => {
      if (!knId) return;
      setConceptGroupsLoading(true);
      try {
        const res = await queryConceptGroups(knId, { limit: -1 });
        setConceptGroups(res?.entries || []);
      } catch (error) {
        console.error('Failed to fetch concept groups:', error);
      } finally {
        setConceptGroupsLoading(false);
      }
    };

    fetchConceptGroups();
    if (id) {
      getObjectDetail();
    }
  }, [id]);

  const getObjectDetail = async () => {
    try {
      const result = await SERVICE.object.getDetail(knId, [id]);
      const data = result?.[0];
      if (!data) return;

      const {
        name,
        tags,
        comment,
        icon,
        color,
        data_properties = [],
        data_source,
        logic_properties = [],
        primary_keys = [],
        display_key = '',
        incremental_key = '',
        concept_groups = [],
      } = data;

      setBasicValue({
        name,
        id,
        tags,
        comment,
        icon,
        color,
        concept_groupIds: concept_groups.map((item) => item.id),
      });
      setDataProperties(data_properties);
      setLogicProperties(logic_properties);
      setDataSource(data_source);
      setPrimaryKeys(primary_keys);
      setDisplayKey(display_key);
      setIncrementalKey(incremental_key);
      setDoneStep(2);
    } catch (error) {
      console.error('getObjectDetail error:', error);
    }
  };

  const goBack = () => history.goBack();
  const onPrev = () => setStepsCurrent((prev) => prev - 1);
  const onNext = () => {
    setDoneStep((prev) => prev + 1);
    setStepsCurrent((prev) => prev + 1);
  };

  const onSubmit = useCallback(
    async (data: { logicProperties: OntologyObjectType.LogicProperty[] }) => {
      const { concept_groupIds = [], ...rest } = basicValue || {};

      const concept_groups = conceptGroups.filter((group) => concept_groupIds.includes(group.id)).map((group) => ({ id: group.id, name: group.name || '' }));

      const dataPropertiesRes = dataProperties.map((item) => ({
        display_name: item.display_name || '',
        name: item.name || '',
        original_name: item.original_name || '',
        type: item.type || '',
        comment: item.comment || '',
        mapped_field: item.mapped_field,
        index_config: item.index_config,
      }));

      const postData: OntologyObjectType.ReqObjectType = {
        branch: 'main',
        ...(rest as OntologyObjectType.BasicInfo)!,
        concept_groups,
        data_properties: dataPropertiesRes,
        logic_properties: data.logicProperties,
        data_source: {
          type: 'data_view',
          id: dataSource?.id || '',
          name: dataSource?.name || '',
        },
        primary_keys: dataProperties.filter((item) => item.primary_key).map((item) => item.name),
        display_key: dataProperties.find((item) => item.display_key)?.name || '',
        incremental_key: dataProperties.find((item) => item.incremental_key)?.name || '',
      };

      setLoading(true);
      try {
        if (isEditPage) {
          await SERVICE.object.updateObjectType(knId, id, postData);
          message.success(intl.get('Global.editSuccess'));
        } else {
          const result = await SERVICE.object.createObjectTypes(knId, [postData]);
          message.success(intl.get('Global.createSuccess'));
          if (!result?.[0]?.id) return;
        }
        goBack();
      } catch (error) {
        console.error('onSubmit error:', error);
      } finally {
        setLoading(false);
      }
    },
    [basicValue, conceptGroups, dataProperties, dataSource, isEditPage, knId, id, message, goBack]
  );

  const title = isEditPage ? basicValue?.name : intl.get('Global.createObjectType');

  const handleBasicNext = () => {
    basicForm.validateFields().then((values) => {
      setBasicValue(values);
      onNext();
    });
  };

  const handleDataPrev = () => {
    dataAttributeRef.current
      .getDataProperties()
      .then((values: { dataProperties: OntologyObjectType.DataProperty[]; dataSource: OntologyObjectType.DataSource }) => {
        setDataProperties(values.dataProperties || []);
        setDataSource(values.dataSource);
        onPrev();
      });
  };

  const handleDataNext = () => {
    dataAttributeRef.current
      .validateFields()
      .then((values: { dataProperties: OntologyObjectType.DataProperty[]; dataSource: OntologyObjectType.DataSource }) => {
        setDataProperties(values.dataProperties || []);
        setDataSource(values.dataSource);
        onNext();
      });
  };

  const handleLogicPrev = () => {
    logicAttributeRef.current.validateFields().then((values: { logicProperties: OntologyObjectType.LogicProperty[] }) => {
      setLogicProperties(values.logicProperties);
      onPrev();
    });
  };

  const handleLogicSave = () => {
    logicAttributeRef.current.validateFields().then((logicValues: { logicProperties: OntologyObjectType.LogicProperty[] }) => {
      onSubmit({ logicProperties: logicValues.logicProperties });
    });
  };

  const stepsContent = useMemo(
    () => ({
      0: {
        content: (
          <BasicInformation
            form={basicForm}
            values={basicValue}
            isEditPage={isEditPage}
            conceptGroups={conceptGroups}
            conceptGroupsLoading={conceptGroupsLoading}
          />
        ),
        prevText: intl.get('Global.prev'),
        nextText: intl.get('Global.next'),
        nextClick: handleBasicNext,
      },
      1: {
        content: (
          <DataAttribute
            dataProperties={dataProperties}
            logicProperties={logicProperties}
            dataSource={dataSource}
            primaryKeys={primaryKeys}
            displayKey={displayKey}
            incrementalKey={incrementalKey}
            basicValue={basicValue}
            ref={dataAttributeRef}
          />
        ),
        prevText: intl.get('Global.prev'),
        nextText: intl.get('Global.next'),
        prevClick: handleDataPrev,
        nextClick: handleDataNext,
      },
      2: {
        content: <LogicAttribute ref={logicAttributeRef} basicValue={basicValue as any} dataProperties={dataProperties} logicProperties={logicProperties} />,
        prevText: intl.get('Global.prev'),
        saveText: intl.get('Global.saveAndExit'),
        prevClick: handleLogicPrev,
        saveClick: handleLogicSave,
      },
    }),
    [
      basicValue,
      basicForm,
      conceptGroups,
      conceptGroupsLoading,
      isEditPage,
      dataProperties,
      logicProperties,
      dataSource,
      primaryKeys,
      displayKey,
      incrementalKey,
      handleBasicNext,
      handleDataPrev,
      handleDataNext,
      handleLogicPrev,
      handleLogicSave,
    ]
  );

  const currentStep = stepsContent[stepsCurrent as keyof typeof stepsContent] as {
    content: React.ReactNode;
    prevText?: string;
    nextText?: string;
    saveText?: string;
    prevClick?: () => void;
    nextClick?: () => void;
    saveClick?: () => void;
  };

  const handleStepChange = (current: number) => {
    if (current <= doneStep) {
      setStepsCurrent(current);
    }
  };

  const headerActions = useMemo(() => {
    const actions: any = {};

    if (currentStep?.prevClick) {
      actions.prev = {
        text: currentStep.prevText,
        onClick: currentStep.prevClick,
        loading,
        disabled: loading,
      };
    }

    if (currentStep?.saveClick) {
      actions.save = {
        text: currentStep.saveText,
        onClick: currentStep.saveClick,
        loading,
        disabled: loading,
      };
    }

    if (currentStep?.nextClick) {
      actions.next = {
        text: currentStep.nextText,
        onClick: currentStep.nextClick,
        loading,
        disabled: loading,
      };
    }

    return Object.keys(actions).length > 0 ? actions : undefined;
  }, [currentStep, loading]);

  return (
    <div className={styles['object-root']}>
      <HeaderSteps title={title} stepsCurrent={stepsCurrent} items={steps} actions={headerActions} />
      <div className={styles['object-content']}>{currentStep?.content}</div>
    </div>
  );
};

export default ObjectCreateAndEdit;

import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Select } from 'antd';
import { getObjectTags } from '@/services/tag';
import locales from './locales'; // 国际化

const TagsSelector: React.FC<any> = (props: any) => {
  const [tagsData, setTagsData] = useState<Array<{ tag: string; count: number }>>([]);
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  useEffect(() => {
    const getAllTags = async (): Promise<void> => {
      const res = await getObjectTags({
        sort: 'tag',
        direction: 'asc',
        limit: -1,
      });
      setTagsData(res.entries.map((item) => ({ tag: item.tag, count: item.count || 0 })));
    };

    if (!props.isOld) getAllTags();
  }, []);

  const handleChange = (val: Array<string>): void => {
    const newVal = val
      .map((i) => i.trim())
      .filter((i) => i)
      .sort();

    if (props.onChange) {
      props.onChange(newVal);
    }
  };

  return (
    <Select
      mode="tags"
      value={props.value}
      onChange={handleChange}
      placeholder={props.placeholder ?? intl.get('TagsSelector.addTags')}
      allowClear
      style={{ width: '100%' }}
    >
      {tagsData &&
        tagsData.map((item) => (
          <Select.Option value={item.tag} key={item.tag}>
            {item.tag}
          </Select.Option>
        ))}
    </Select>
  );
};

export default TagsSelector;

export const tagsSelectorValidator = (_rule: any, value: Array<string> | undefined) => {
  if (value && value.length > 5) return Promise.reject(new Error(intl.get('TagsSelector.tagQuantityLimitInfo')));
  if (value && value.length) {
    const regex = /^[^/:?\\"<>|：?""？！《》,#[]{}%&*$^!=.'']*$/;
    value.forEach((tag) => {
      if (tag.length > 40) return Promise.reject(new Error(intl.get('TagsSelector.tagLengthLimitInfo')));
      if (!regex.test(tag)) return Promise.reject(new Error(intl.get('TagsSelector.tagParticularCharacter')));
    });
  }
  return Promise.resolve();
};

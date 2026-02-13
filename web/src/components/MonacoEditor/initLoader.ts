import { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';

let monacoLoaderInited = false;

export const initMonacoLoader = () => {
  if (monacoLoaderInited) return;
  const monacoConfig = {
    monaco,
  };

  loader.config(monacoConfig);

  const amdRequire = (window as any)?.require;
  if (typeof amdRequire?.config === 'function') {
    amdRequire.config(monacoConfig);
  }

  monacoLoaderInited = true;
};

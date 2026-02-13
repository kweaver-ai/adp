/**
 * @description 代码编辑器组件
 */
import { useRef, forwardRef, useImperativeHandle } from 'react';
import Editor from '@monaco-editor/react';
import Compound from './Compound';
import { initMonacoLoader } from './initLoader';
import type { EditorProps } from '@monaco-editor/react';

initMonacoLoader();
const CustomMonacoEditor: React.FC<any> = forwardRef((props, ref) => {
  const { onChange: propOnChange, ...otherProps } = props;
  const monacoRef = useRef<any>();
  const editorRef = useRef<any>();

  useImperativeHandle(ref, () => ({
    getMonacoInstance: () => monacoRef.current,
    getEditorInstance: () => editorRef.current,
  }));

  const handleMount: EditorProps['onMount'] = (editor, monaco) => {
    monacoRef.current = monaco;
    editorRef.current = editor;
  };

  const handelChange: EditorProps['onChange'] = (value, ev) => {
    propOnChange?.(value, ev);
  };

  return (
    <Editor height="100%" onMount={handleMount} onChange={handelChange} beforeMount={() => editorRef.current?.getMonacoInstance()?.dispose()} {...otherProps} />
  );
});

type CustomMonacoEditorProps = typeof CustomMonacoEditor & {
  Compound: typeof Compound;
};

const MonacoEditor = Object.assign(CustomMonacoEditor, {
  Compound,
}) as CustomMonacoEditorProps;

export default MonacoEditor;

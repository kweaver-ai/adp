/**
 * @description Monaco Editor 动态加载包装器
 * 用于实现代码分割，在首屏不加载 Monaco 库
 * 这样可以减少主应用 chunk 的体积，只在需要时才加载
 */
import { lazy, Suspense, forwardRef } from 'react';
import { Spin } from 'antd';

// 动态导入 Monaco Editor 主组件
const LazyMonacoEditor = lazy(() => import('./index'));

// 动态导入 Monaco Editor Compound 组件
const LazyMonacoEditorCompound = lazy(() => import('./Compound'));

interface LoadingFallbackProps {
  height?: number | string;
  width?: number | string;
}

/**
 * 加载中的 Fallback UI
 */
const LoadingFallback: React.FC<LoadingFallbackProps> = ({ height = '100%', width = '100%' }) => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height,
      width,
      minHeight: '200px',
    }}
  >
    <Spin tip="Loading Editor..." />
  </div>
);

/**
 * 为 Monaco Editor 提供 Loading 和 Error 状态的包装器
 */
const MonacoEditorWrapper = forwardRef<any, any>((props: any, ref: any) => {
  const { height = '100%', width = '100%', ...otherProps } = props;
  return (
    <Suspense fallback={<LoadingFallback height={height} width={width} />}>
      <LazyMonacoEditor {...otherProps} height={height} width={width} ref={ref} />
    </Suspense>
  );
});

/**
 * 为 Monaco Editor Compound 提供 Loading 和 Error 状态的包装器
 */
const MonacoEditorCompoundWrapper = forwardRef<any, any>((props: any, ref: any) => {
  const { height = '100%', width = '100%', ...otherProps } = props;
  return (
    <Suspense fallback={<LoadingFallback height={height} width={width} />}>
      <LazyMonacoEditorCompound {...otherProps} height={height} width={width} ref={ref} />
    </Suspense>
  );
});

MonacoEditorWrapper.displayName = 'MonacoEditorWrapper';
MonacoEditorCompoundWrapper.displayName = 'MonacoEditorCompoundWrapper';

// 保持与原始组件相同的结构
export const LazyMonacoEditorComponent = Object.assign(MonacoEditorWrapper, {
  Compound: MonacoEditorCompoundWrapper,
});

export default LazyMonacoEditorComponent;

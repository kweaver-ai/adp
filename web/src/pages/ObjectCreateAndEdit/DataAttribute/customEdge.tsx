import { memo } from 'react';
import { CloseOutlined } from '@ant-design/icons';
import { BaseEdge, EdgeLabelRenderer, getBezierPath, useReactFlow, type EdgeProps } from '@xyflow/react';
import styles from './index.module.less';

export const CustomEdge = memo((props: EdgeProps) => {
  const { selected, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, markerEnd, id } = props;
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });
  const reactFlowInstance = useReactFlow();

  const deleteEdge = () => {
    reactFlowInstance.deleteElements({ edges: [{ id }] });
  };

  return (
    <>
      <BaseEdge path={edgePath} style={{ ...style, pointerEvents: 'all', stroke: selected ? '#555' : '#b1b1b1' }} />
      <EdgeLabelRenderer>
        {selected && (
          <div
            className={styles['edge-btn']}
            style={{
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
            }}
            onClick={deleteEdge}
          >
            <CloseOutlined style={{ color: '#fff', fontSize: 10 }} />
          </div>
        )}
      </EdgeLabelRenderer>
    </>
  );
});

export default CustomEdge;

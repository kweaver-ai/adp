import { memo } from 'react';
import { CloseOutlined } from '@ant-design/icons';
import { BaseEdge, EdgeLabelRenderer, getBezierPath, useReactFlow, type EdgeProps } from '@xyflow/react';
import styles from './index.module.less';

export const CustomEdge = memo((props: EdgeProps) => {
  const { selected, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, id, data } = props;
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

  const isHovered = data?.isHovered || false;
  const strokeColor = selected || isHovered ? '#000' : '#b1b1b1';

  return (
    <>
      <BaseEdge path={edgePath} style={{ ...style, pointerEvents: 'all', stroke: strokeColor, transition: 'all 0.2s ease' }} />
      <EdgeLabelRenderer>
        {isHovered && (
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

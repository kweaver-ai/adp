import { createContext, useContext } from 'react';

export const HoveredEdgeIdContext = createContext<string | null>(null);

export const useHoveredEdgeId = () => useContext(HoveredEdgeIdContext);

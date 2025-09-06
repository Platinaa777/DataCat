import { DataSource } from '../alerting';

export enum VisualizationType {
  LINE = 'line',
  BAR = 'bar',
  PIE = 'pie',
  GAUGE = 'gauge',
  TABLE = 'table',
  UNKNOWN = 'unknown',
}

export type Position = 'top' | 'bottom' | 'left' | 'right';

export interface BasicVisaulizationOptions {
  legend?: {
    enabled?: boolean;
    position?: Position;
  };
  title?: {
    enabled?: boolean;
    text?: string;
  };
  tooltip?: {
    enabled?: boolean;
  };
}

export type LineVisualizationOptions = BasicVisaulizationOptions;
export type BarVisualizationOptions = BasicVisaulizationOptions;
export type PieVisualizationOptions = BasicVisaulizationOptions;

export type VisualizationSettings =
  | LineVisualizationOptions
  | BarVisualizationOptions
  | PieVisualizationOptions;

export interface Layout {
  x: number;
  y: number;
  rows: number;
  cols: number;
}

export interface Panel {
  id: string;
  title: string;
  query: string;
  dataSource?: DataSource;
  layout: Layout;
  visualizationType?: VisualizationType;
  visualizationSettings?: VisualizationSettings;
}

export interface LineStyle {
  lineWidth: number;
}

export const decodeLayout = (encoded: string | undefined): Layout => {
  if (encoded) {
    try {
      const layout = JSON.parse(encoded);
      return layout;
    } catch {
      return { x: 0, y: 0, cols: 5, rows: 3 };
    }
  }

  return {
    x: 0,
    y: 0,
    cols: 5,
    rows: 3,
  };
};

export const encodeLayout = (layout: Layout): string => {
  return JSON.stringify(layout);
};

export const decodeVisualizationSettings = (
  settings: string | undefined,
): VisualizationSettings => {
  if (!settings) {
    return {};
  }
  try {
    return JSON.parse(settings);
  } catch {
    return {};
  }
};

export const encodeVisualizationSettings = (
  settings: VisualizationSettings | undefined,
): string => {
  return JSON.stringify(settings);
};

export const encodeVisualizationType = (
  type: VisualizationType | undefined,
): number => {
  switch (type) {
    case VisualizationType.LINE:
      return 1;
    case VisualizationType.BAR:
      return 2;
    case VisualizationType.PIE:
      return 3;
    default:
      return 4;
  }
};

export const decodeVisualizationType = (
  type: string | undefined,
): VisualizationType => {
  switch (type) {
    case 'Graph':
      return VisualizationType.LINE;
    case 'Table':
      return VisualizationType.BAR;
    case 'Pie Chart':
      return VisualizationType.PIE;
    default:
      return VisualizationType.UNKNOWN;
  }
};

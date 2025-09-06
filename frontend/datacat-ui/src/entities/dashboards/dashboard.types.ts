import { Panel } from './panel.types';

export interface Dashboard {
  id: string;
  name: string;
  description: string;
  panels: Panel[];
  createdAt?: Date;
  lastUpdatedAt?: Date;
}

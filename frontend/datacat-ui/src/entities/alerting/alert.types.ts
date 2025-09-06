export enum AlertStatus {
  OK = 'Ok',
  FIRING = 'Firing',
  ERROR = 'Error',
  MUTED = 'Muted',
}

export interface Alert {
  id: string;
  template: string;
  description: string;
  query: string;
  status: AlertStatus;
  dataSourceId: string;
  notificationGroupId: string;
  prevExecutionTime?: Date;
  nextExecutionTime?: Date;
  notificationTriggerPeriod: string;
  executionInterval: string;
}

export type AlertExpanded = Record<string, unknown>;

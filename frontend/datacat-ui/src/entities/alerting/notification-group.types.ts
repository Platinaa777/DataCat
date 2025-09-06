import { NotificationChannel } from './notification-channel.types';

export enum NotificationTemplateSyntax {
  MARKDOWN = 'markdown',
}

export interface NotificationTemplate {
  syntax: NotificationTemplateSyntax;
  template: string;
}

export interface NotificationGroup {
  id: string;
  name: string;
  notificationChannelsIds: string[];
}

export interface NotificationGroupExpanded {
  id: string;
  name: string;
  notificationChannels: NotificationChannel[];
}

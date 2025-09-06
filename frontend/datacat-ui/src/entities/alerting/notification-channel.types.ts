

export interface EmailSettings {
  DestinationEmail: string;
  SmtpServer: string;
  Port: number;
  PasswordPath: string;
};

export interface WebhookSettings {
  Url: string;
}

export interface TelegramSettings {
  TelegramTokenPath: string;
  ChatId: string;
}

export enum NotificationChannelDriver {
  EMAIL = 'email',
  WEBHOOK = 'webhook',
  TELEGRAM = 'telegram',
};

export type NotificationChannelSettings = EmailSettings | WebhookSettings | TelegramSettings;

export interface NotificationChannel {
  id: number;
  driver: NotificationChannelDriver;
  settings: NotificationChannelSettings;
}

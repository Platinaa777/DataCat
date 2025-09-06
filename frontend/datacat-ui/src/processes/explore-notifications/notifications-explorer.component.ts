import { Component } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { DividerModule } from 'primeng/divider';
import { Panel } from 'primeng/panel';
import { TabsModule } from 'primeng/tabs';
import {
NotificationChannelDriver,
NotificationGroup,
} from '../../entities';
import { CreateNotificationGroupButtonComponent } from '../../features/alerting/create-notification-group';
import { NotificationGroupsListComponent } from '../../features/alerting/notification-groups-list';

@Component({
  standalone: true,
  selector: './datacat-notification-channels-crud',
  templateUrl: './notifications-explorer.component.html',
  styleUrl: './notifications-explorer.component.scss',
  imports: [
    Panel,
    TabsModule,
    DividerModule,
    ButtonModule,
    CreateNotificationGroupButtonComponent,
    NotificationGroupsListComponent,
  ],
})
export class NotificationsExplorerComponent {
  protected NotificationChannelDriver = NotificationChannelDriver;

  protected channelsGroups?: NotificationGroup[];
}

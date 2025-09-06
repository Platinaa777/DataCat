import { Component,Input } from '@angular/core';
import { PanelModule } from 'primeng/panel';
import { EditNotificationGroupFormComponent } from '../../features/alerting/edit-notification-group';

@Component({
  standalone: true,
  selector: 'datacat-manage-notification-group',
  templateUrl: './manage-notification-group.component.html',
  styleUrl: './manage-notification-group.component.scss',
  imports: [
    PanelModule,
    EditNotificationGroupFormComponent,
  ],
})
export class ManageNotificationGroupComponent {
  @Input() groupId?: string;
}

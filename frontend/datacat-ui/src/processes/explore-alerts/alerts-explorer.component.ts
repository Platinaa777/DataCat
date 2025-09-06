import { Component } from '@angular/core';
import { PanelModule } from 'primeng/panel';
import { AlertsCountsByStatusComponent } from '../../features/alerting/alerts-counts-by-status/alerts-counts-by-status.component';
import { AlertsListComponent } from '../../features/alerting/alerts-list/alerts-list.component';
import { CreateAlertButtonComponent } from '../../features/alerting/create-alert';

@Component({
  standalone: true,
  selector: 'datacat-alerts-explore',
  templateUrl: './alerts-explorer.component.html',
  styleUrl: './alerts-explorer.component.scss',
  imports: [
    AlertsCountsByStatusComponent,
    AlertsListComponent,
    PanelModule,
    CreateAlertButtonComponent,
  ],
})
export class AlertsExplorerComponent {
}

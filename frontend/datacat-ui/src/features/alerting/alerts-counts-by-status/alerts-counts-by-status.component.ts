import { Component,OnInit } from '@angular/core';
import { TagModule } from 'primeng/tag';
import { TooltipModule } from 'primeng/tooltip';
import { ApiService } from '../../../shared/services/datacat-generated-client';
import { ToastLoggerService } from '../../../shared/services/toast-logger.service';
import { AlertStatus } from '../../../entities';
import { AlertsCountsByStatus } from './alerts-counts-by-status.types';

@Component({
  standalone: true,
  selector: 'datacat-alerts-counts-by-status',
  templateUrl: './alerts-counts-by-status.component.html',
  styleUrl: './alerts-counts-by-status.component.scss',
  imports: [TagModule, TooltipModule],
})
export class AlertsCountsByStatusComponent implements OnInit {
  protected alertsCountsByStatus?: AlertsCountsByStatus;

  constructor(private api: ApiService, private logger: ToastLoggerService) {}

  ngOnInit() {
    this.loadAlertsCountsByStatus();
  }

  protected loadAlertsCountsByStatus() {
    this.api.getApiV1AlertGetCounters().subscribe({
      next: counters => {
        this.alertsCountsByStatus = new Map(counters.map(counter => [counter.status as AlertStatus, counter.count ?? 0]));
      },
      error: error => this.logger.error(error),
    });
  }

  protected getSeverityForStatus(
    status: AlertStatus,
  ): 'success' | 'danger' | 'secondary' | 'info' {
    {
      switch (status) {
        case AlertStatus.OK: {
          return 'success';
        }
        case AlertStatus.FIRING:
        case AlertStatus.ERROR: {
          return 'danger';
        }
        case AlertStatus.MUTED: {
          return 'secondary';
        }
        default: {
          return 'secondary';
        }
      }
    }
  }
}

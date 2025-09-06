import { DatePipe,NgIf } from "@angular/common";
import { Component,EventEmitter,Input,Output } from '@angular/core';
import { Button } from "primeng/button";
import { TableModule } from "primeng/table";
import { Tag } from "primeng/tag";
import { AppDialogService } from "../../../shared/services/app-dialog.service";
import { LogEntry } from "../../../shared/services/datacat-generated-client";
import { sanitizeMessage } from "../../../shared/utils/sanitizeMessage";
import { LogsDetailsComponent } from "../logs-details/logs-details.component";

@Component({
    selector: 'app-logs-list',
    standalone: true,
    imports: [
        TableModule,
        DatePipe,
        Tag,
        NgIf,
        Button
    ],
    templateUrl: './logs-list.component.html',
    styleUrl: './logs-list.component.scss'
})
export class LogsListComponent {
    @Input() logs: LogEntry[] = [];
    @Input() pagination: any;
    @Output() pageChange = new EventEmitter<number>();
    @Output() sortChange = new EventEmitter<{ field: string; ascending: boolean }>();

    currentSortField = '';
    sortAscending = false;
    protected readonly sanitizeMessage = sanitizeMessage;

    constructor(
        private appDialogService: AppDialogService,
    ) {
    }

    onPageChange(event: any) {
        const newPage = (event.first / event.rows) + 1;
        this.pageChange.emit(newPage);
    }


    onSort(field: string) {
        if (this.currentSortField === field) {
            this.sortAscending = !this.sortAscending;
        } else {
            this.currentSortField = field;
            this.sortAscending = true;
        }
        this.sortChange.emit({field: this.currentSortField, ascending: this.sortAscending});
    }

    getSeverityClass(severity: string): 'success' | 'info' | 'warn' | 'danger' {
        switch (severity?.toLowerCase()) {
            case 'error':
                return 'danger';
            case 'warning':
                return 'warn';
            case 'info':
                return 'info';
            default:
                return 'success';
        }
    }

    showDetails(additionalFields: any) {
        this.appDialogService.showDialog(LogsDetailsComponent,
            'Additional Fields',
            additionalFields || {}
        )
    }
}

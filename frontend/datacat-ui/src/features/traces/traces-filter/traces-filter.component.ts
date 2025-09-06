import { Component,EventEmitter,Input,Output } from '@angular/core';
import { FormsModule } from "@angular/forms";
import { Button } from "primeng/button";
import { Calendar } from "primeng/calendar";
import { DropdownModule } from "primeng/dropdown";
import { InputText } from "primeng/inputtext";
import { catchError,tap } from "rxjs";
import { ApiService,ISearchTracesRequest } from "../../../shared/services/datacat-generated-client";
import { ToastLoggerService } from "../../../shared/services/toast-logger.service";

@Component({
    selector: 'app-traces-filter',
    standalone: true,
    imports: [
        DropdownModule,
        FormsModule,
        Calendar,
        InputText,
        Button
    ],
    templateUrl: './traces-filter.component.html',
    styleUrl: './traces-filter.component.scss'
})
export class TracesFilterComponent {
    @Output() filterChange = new EventEmitter<ISearchTracesRequest>();
    services: string[] = [];
    operations: string[] = [];

    constructor(
        private apiService: ApiService,
        private toastLoggerService: ToastLoggerService) {
    }

    private _currentFilter!: ISearchTracesRequest;

    get currentFilter(): ISearchTracesRequest {
        return this._currentFilter;
    }

    @Input()
    set currentFilter(value: ISearchTracesRequest) {
        this._currentFilter = {...value};
        if (value.dataSourceName) {
            this.loadServices();
        }
    }

    loadServices() {
        if (!this.currentFilter.dataSourceName) return;

        this.apiService.getApiV1TracesServices(this.currentFilter.dataSourceName)
            .pipe(
                tap(services => this.services = services),
                catchError(err => {
                    this.toastLoggerService.error(err.message);
                    return err;
                })
            ).subscribe();
    }

    loadOperations() {
        if (!this.currentFilter.dataSourceName || !this.currentFilter.service) return;

        this.apiService.getApiV1TracesOperations(
            this.currentFilter.dataSourceName,
            this.currentFilter.service)
            .pipe(
                tap(operations => this.operations = operations),
                catchError(err => {
                    this.toastLoggerService.error(err.message);
                    return err;
                })
            ).subscribe();
    }

    onServiceChange() {
        this.currentFilter.operation = undefined;
        this.loadOperations();
    }

    applyFilters() {
        this.filterChange.emit(this.currentFilter);
    }
}

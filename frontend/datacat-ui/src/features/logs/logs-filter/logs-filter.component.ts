import { NgForOf } from "@angular/common";
import { Component,EventEmitter,Input,Output } from '@angular/core';
import { FormsModule } from "@angular/forms";
import { ButtonDirective } from "primeng/button";
import { Calendar } from "primeng/calendar";
import { DropdownModule } from "primeng/dropdown";
import { InputText } from "primeng/inputtext";
import { ISearchLogsRequest } from "../../../shared/services/datacat-generated-client";

@Component({
    selector: 'app-logs-filter',
    standalone: true,
    imports: [
        InputText,
        DropdownModule,
        Calendar,
        FormsModule,
        NgForOf,
        ButtonDirective
    ],
    templateUrl: './logs-filter.component.html',
    styleUrl: './logs-filter.component.scss'
})
export class LogsFilterComponent {
    @Input() initialFilter!: ISearchLogsRequest;
    @Output() filterChange = new EventEmitter<ISearchLogsRequest>();

    filter: ISearchLogsRequest;
    customFilters: { key: string; value: string }[] = [];
    severityLevels = [
        {label: 'Error', value: 'Error'},
        {label: 'Warning', value: 'Warning'},
        {label: 'Info', value: 'Information'},
        {label: 'Debug', value: 'Debug'}
    ];

    showCustomFilters = false;

    constructor() {
        this.filter = {...this.initialFilter};
        if (this.filter.customFilters) {
            this.customFilters = Object.entries(this.filter.customFilters)
                .map(([key, value]) => ({key, value}));
        }
    }

    addCustomFilter() {
        this.customFilters.push({key: '', value: ''});
    }

    removeCustomFilter(index: number) {
        this.customFilters.splice(index, 1);
    }

    applyFilters() {
        const customFilters = this.customFilters.reduce((acc, curr) => {
            if (curr.key && curr.value) acc[curr.key] = curr.value;
            return acc;
        }, {} as Record<string, string>);

        this.filterChange.emit({
            ...this.filter,
            customFilters: Object.keys(customFilters).length ? customFilters : undefined
        });
    }
}

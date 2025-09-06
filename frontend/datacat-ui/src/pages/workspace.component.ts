import { Component,computed } from '@angular/core';
import { Router,RouterModule } from '@angular/router';

import { Location } from '@angular/common';
import { FormsModule } from "@angular/forms";
import { MenuItem,TreeNode } from 'primeng/api';
import { BreadcrumbModule } from 'primeng/breadcrumb';
import { ButtonModule } from 'primeng/button';
import { SelectModule } from 'primeng/select';
import { SplitterModule } from 'primeng/splitter';
import { Toast } from 'primeng/toast';
import { TreeModule,TreeNodeSelectEvent } from 'primeng/tree';
import { ActivityBreadcrumbComponent } from '../features/workspace/activity-breadcrumb';
import * as urls from '../shared/common/urls';
import { NamespaceService } from "../shared/services/namespace.service";

interface TreeMenuNodeData {
    url?: string;
}

@Component({
    standalone: true,
    selector: 'datacat-main-page',
    templateUrl: './workspace.component.html',
    styleUrl: './workspace.component.scss',
    imports: [
        RouterModule,
        SplitterModule,
        BreadcrumbModule,
        TreeModule,
        SelectModule,
        ButtonModule,
        ActivityBreadcrumbComponent,
        Toast,
        FormsModule,


    ],
})
export class WorkspaceComponent {
    namespaces = computed(() => this.namespaceService.namespaces());
    selectedNamespace = computed(() => this.namespaceService.currentNamespace());

    protected activityPathItems: MenuItem[] = [];
    protected readonly treeMenuRootNode: TreeNode<TreeMenuNodeData>[] = [
        {
            type: 'url',
            icon: 'pi pi-megaphone',
            label: 'Alerts',
            data: {
                url: urls.ALERTS_EXPLORER_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-envelope',
            label: 'Notifications',
            data: {
                url: urls.NOTIFICATIONS_EXPLORER_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-table',
            label: 'Dashboards',
            data: {
                url: urls.DASHBOARDS_EXPLORER_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-crown',
            label: 'Admin',
            data: {
                url: urls.ADMIN_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-database',
            label: 'Data Sources',
            data: {
                url: urls.DATA_SOURCES_EXPLORER_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-file',
            label: 'Logs',
            data: {
                url: urls.LOGS_EXPLORER_URL,
            },
        },
        {
            type: 'url',
            icon: 'pi pi-sliders-h',
            label: 'Traces',
            data: {
                url: urls.TRACES_EXPLORER_URL,
            },
        },
    ];

    constructor(
        private router: Router,
        private location: Location,
        private namespaceService: NamespaceService
    ) {

    }

    onNamespaceChange(ns: any): void {
        this.namespaceService.setCurrentNamespace(ns.value);
        this.router.navigate(['/dashboards']).then(() => window.location.reload());
    }

    protected onTreeMenuNodeSelect(event: TreeNodeSelectEvent) {
        if (event.node.data && event.node.data.url) {
            this.router.navigate([event.node.data.url]);
        }
    }

    protected goToPreviousActivity() {
        this.location.back();
    }

    protected goToNextActivity() {
        this.location.forward();
    }

    protected openSettings() {
        this.router.navigate(['settings']);
    }
}

import { Component,Input } from '@angular/core';
import {
FormControl,
FormGroup,
ReactiveFormsModule,
Validators,
} from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { ChipModule } from 'primeng/chip';
import { DialogModule } from 'primeng/dialog';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { InputMaskModule } from 'primeng/inputmask';
import { InputTextModule } from 'primeng/inputtext';
import { SelectModule } from 'primeng/select';
import { TextareaModule } from 'primeng/textarea';
import { finalize } from 'rxjs';
import { DataSource,NotificationGroup } from '../../../entities';
import { ApiService } from '../../../shared/services/datacat-generated-client';
import { ToastLoggerService } from '../../../shared/services/toast-logger.service';
import { DataSourceSelectComponent } from '../../../shared/ui/data-source-select/data-source-select.component';
import { NotificationGroupSelectComponent } from '../../../shared/ui/notification-group-select/notification-group-select.component';

@Component({
  standalone: true,
  selector: './datacat-edit-alert-form',
  templateUrl: './edit-alert-form.component.html',
  styleUrl: './edit-alert-form.component.scss',
  imports: [
    ButtonModule,
    DialogModule,
    ReactiveFormsModule,
    InputTextModule,
    TextareaModule,
    SelectModule,
    ChipModule,
    InputGroupModule,
    InputGroupAddonModule,
    DataSourceSelectComponent,
    NotificationGroupSelectComponent,
    InputMaskModule,
  ],
})
export class EditAlertFormComponent {
  private _alertId?: string;

  @Input() public set alertId(id: string) {
    this._alertId = id;
    this.refresh();
  }

  protected isSavingInitiated = false;

  protected addTagControl = new FormControl<string>('');

  protected editForm = new FormGroup({
    description: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    template: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    query: new FormControl<string>('', [Validators.required]),
    dataSourceId: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    notificationGroupName: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    notificationTriggerPeriod: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    executionInterval: new FormControl<string | undefined>(undefined, {
      nonNullable: false,
      validators: Validators.required,
    }),
    tags: new FormControl<string[]>([]),
  });

  protected get editFormTags(): string[] {
    return this.editForm.get('tags')?.value || [];
  }

  protected refresh() {
    if (!this._alertId) {
      return;
    }

    this.editForm.disable();
    this.apiService
      .getApiV1Alert(this._alertId)
      .pipe(
        finalize(() => {
          this.editForm.enable();
        }),
      )
      .subscribe({
        next: (data) => {
          this.editForm.setValue({
            description: data.description || '',
            template: data.template || '',
            query: data.rawQuery || '',
            dataSourceId: data.dataSource?.id || '',
            notificationGroupName: data.notificationChannelGroup?.name || '',
            notificationTriggerPeriod: data.waitTimeBeforeAlerting || '',
            executionInterval: data.repeatInterval || '',
            tags: data.tags || [],
          });
        },
        error: (e) => {
          this.loggerService.error(e);
        },
      });
  }

  protected addEditFormTag() {
    const tag = this.addTagControl.value;
    const previousTags = this.editForm.get('tags')?.value || [];
    if (tag && !previousTags.includes(tag)) {
      this.editForm.get('tags')?.setValue([...previousTags, tag]);
    }
  }

  protected removeEditFormTag(tag: string) {
    const previousTags = this.editFormTags;
    this.editForm.get('tags')?.setValue(previousTags.filter((t) => t !== tag));
  }

  protected dataSources: DataSource[] = [];
  protected notificationGroups: NotificationGroup[] = [];

  constructor(
    private apiService: ApiService,
    private loggerService: ToastLoggerService,
  ) {}

  protected saveAlert() {
    this.editForm.markAllAsTouched();
    this.editForm.updateValueAndValidity();

    if (this.editForm.invalid) {
      return;
    }

    this.isSavingInitiated = true;

    const rawForm = this.editForm.getRawValue();
    const request: any = {
      description: rawForm.description,
      template: rawForm.template,
      rawQuery: rawForm.query,
      dataSourceId: rawForm.dataSourceId,
      notificationChannelGroupName: rawForm.notificationGroupName,
      waitTimeBeforeAlerting: rawForm.notificationTriggerPeriod,
      repeatInterval: rawForm.executionInterval,
      tags: rawForm.tags,
    };

    this.apiService
      .postApiV1AlertAdd(request)
      .pipe(
        finalize(() => {
          this.isSavingInitiated = false;
        }),
      )
      .subscribe({
        next: () => {
          this.loggerService.success('Saved');
        },
        error: (e) => {
          this.loggerService.error(e);
        },
      });
  }
}

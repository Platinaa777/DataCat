import { Component } from '@angular/core';
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
  selector: './datacat-create-alert-button',
  templateUrl: './create-alert-button.component.html',
  styleUrl: './create-alert-button.component.scss',
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
export class CreateAlertButtonComponent {
  protected isCreationDialogVisible = false;
  protected isCreationInitiated = false;

  protected addTagControl = new FormControl<string>('');

  protected creationForm = new FormGroup({
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

  protected get creationFormTags(): string[] {
    return this.creationForm.get('tags')?.value || [];
  }

  protected addCreationFormTag() {
    const tag = this.addTagControl.value;
    const previousTags = this.creationForm.get('tags')?.value || [];
    if (tag && !previousTags.includes(tag)) {
      this.creationForm.get('tags')?.setValue([...previousTags, tag]);
    }
  }

  protected removeCreationFormTag(tag: string) {
    const previousTags = this.creationFormTags;
    this.creationForm
      .get('tags')
      ?.setValue(previousTags.filter((t) => t !== tag));
  }

  protected dataSources: DataSource[] = [];
  protected notificationGroups: NotificationGroup[] = [];

  constructor(
    private apiService: ApiService,
    private loggerService: ToastLoggerService,
  ) {}

  protected createAlert() {
    if (this.creationForm.invalid) {
      return;
    }

    this.isCreationInitiated = true;

    const rawForm = this.creationForm.getRawValue();
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
          this.isCreationInitiated = false;
        }),
      )
      .subscribe({
        next: () => {
          this.loggerService.success('Alert created');
          this.isCreationDialogVisible = false;
        },
        error: (e) => {
          this.loggerService.error(e);
        },
      });
  }
}

import { Component,Input } from '@angular/core';
import { FormGroup,ReactiveFormsModule } from "@angular/forms";
import { Checkbox } from "primeng/checkbox";
import { InputNumber } from "primeng/inputnumber";
import { InputText } from "primeng/inputtext";

@Component({
    selector: 'app-elasticsearch-settings-form',
    standalone: true,
    imports: [
        InputNumber,
        InputText,
        ReactiveFormsModule,
        Checkbox
    ],
    templateUrl: './elasticsearch-settings-form.component.html',
    styleUrl: './elasticsearch-settings-form.component.scss'
})
export class ElasticsearchSettingsFormComponent {
    @Input() form!: FormGroup;

}

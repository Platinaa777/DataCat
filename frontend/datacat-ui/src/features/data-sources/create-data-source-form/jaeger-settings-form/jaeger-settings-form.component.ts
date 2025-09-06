import { Component,Input } from '@angular/core';
import { FormGroup,ReactiveFormsModule } from "@angular/forms";
import { InputText } from "primeng/inputtext";
import { Select } from "primeng/select";

@Component({
    selector: 'app-jaeger-settings-form',
    standalone: true,
    imports: [
        Select,
        ReactiveFormsModule,
        InputText
    ],
    templateUrl: './jaeger-settings-form.component.html',
    styleUrl: './jaeger-settings-form.component.scss'
})
export class JaegerSettingsFormComponent {
    @Input() form!: FormGroup;
    @Input() authType!: string | undefined | null;

    authTypes = ['None', 'Basic', 'Bearer'];

    onAuthTypeChange() {
        const authType = this.form.value.authType;

        if (authType === 'None') {
            this.form.patchValue({
                username: null,
                password: null,
                authToken: null
            });
        }
    }
}

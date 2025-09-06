import { provideHttpClient,withInterceptors } from '@angular/common/http';
import { APP_INITIALIZER,ApplicationConfig,provideZoneChangeDetection,} from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter,withComponentInputBinding } from '@angular/router';
import { MessageService } from 'primeng/api';
import { providePrimeNG } from 'primeng/config';
import { DialogService } from 'primeng/dynamicdialog';
import { ThemeSelectionService } from '../features/appearence/select-theme/select-theme.service';
import { ROUTES } from '../pages/workspace.routes';
import { apiInterceptor } from '../shared/interceptors/api.interceptor';
import { authInterceptor } from '../shared/interceptors/auth.interceptor';
import { namespaceInterceptor } from "../shared/interceptors/namespace.interceptor";
import { PRIMENG_CONFIG } from '../shared/primeng/primeng.config';
import { ApiService } from '../shared/services/datacat-generated-client';
import { NamespaceService } from "../shared/services/namespace.service";
import { UserService } from "../shared/services/user.service";

export const APP_CONFIG: ApplicationConfig = {
    providers: [
        provideZoneChangeDetection({eventCoalescing: true}),
        provideRouter(ROUTES, withComponentInputBinding()),
        provideAnimationsAsync(),
        provideHttpClient(withInterceptors([apiInterceptor, namespaceInterceptor, authInterceptor])),
        providePrimeNG(PRIMENG_CONFIG),
        DialogService,
        MessageService,
        ApiService,
        ThemeSelectionService,
        NamespaceService,
        UserService,
        {
            provide: APP_INITIALIZER,
            useFactory: (themeService: ThemeSelectionService) => {
                return () => themeService.loadSavedTheme();
            },
            deps: [ThemeSelectionService],
            multi: true,
        },
    ],
};

import { bootstrapApplication } from '@angular/platform-browser';
import { WorkspaceComponent } from '../pages';
import { APP_CONFIG } from './app.config';

bootstrapApplication(WorkspaceComponent, APP_CONFIG).catch((err) =>
  console.error(err),
);

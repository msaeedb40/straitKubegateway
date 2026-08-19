import { ApplicationConfig, provideZonelessChangeDetection, provideAppInitializer, inject, ErrorHandler } from '@angular/core';
import { provideRouter, withComponentInputBinding, withViewTransitions } from '@angular/router';
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';
import { routes } from './app.routes';
import { ConfigService } from './core/config/config.service';
import { apiInterceptor } from './core/api/api.interceptor';
import { authInterceptor } from './core/auth/auth.interceptor';
import { GlobalErrorHandler } from './core/errors/global-error-handler';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZonelessChangeDetection(),
    provideHttpClient(
      withFetch(),
      withInterceptors([apiInterceptor, authInterceptor])
    ),
    provideRouter(routes, withComponentInputBinding(), withViewTransitions()),
    { provide: ErrorHandler, useClass: GlobalErrorHandler },
    provideAppInitializer(async () => {
      const cfg = inject(ConfigService);
      await cfg.loadConfig();
    })
  ]
};

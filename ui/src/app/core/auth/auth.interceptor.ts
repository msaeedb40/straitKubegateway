import { inject } from '@angular/core';
import { HttpInterceptorFn } from '@angular/common/http';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const tokens = authService.tokens();

  if (tokens?.accessToken) {
    const cloned = req.clone({
      setHeaders: {
        Authorization: `Bearer ${tokens.accessToken}`
      }
    });
    return next(cloned);
  }

  return next(req);
};

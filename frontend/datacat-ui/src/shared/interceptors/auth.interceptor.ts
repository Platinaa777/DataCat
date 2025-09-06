import { HttpErrorResponse,HttpInterceptorFn } from '@angular/common/http';
import { catchError,throwError } from 'rxjs';

let loginStarted = false;

export const authInterceptor: HttpInterceptorFn = (req, next) => {
    if (!req.url.startsWith('/api/')) return next(req);
    return next(req.clone({ withCredentials: true })).pipe(
        catchError((error: HttpErrorResponse) => {
            if (error.status === 401 && !loginStarted && !req.url.includes('/user/login')) {
                loginStarted = true;
                window.location.assign('/api/v1/user/login-code-flow');
            }
            return throwError(() => error);
        })
    );
};

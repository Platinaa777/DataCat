import { HttpInterceptorFn } from '@angular/common/http';
import { environment } from '../env/environment';

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
    if (!req.url.startsWith('/api/')) return next(req);
    return next(req.clone({ url: `${environment.apiUrl}${req.url}` }));
};

import { HttpInterceptorFn } from '@angular/common/http';
import { getCurrentNamespaceFromLocalStorage } from '../utils/getCurrentNamespaceFromLocalStorage';

export const namespaceInterceptor: HttpInterceptorFn = (req, next) => {
    const namespace = getCurrentNamespaceFromLocalStorage();
    if (!req.url.startsWith('/api/') || !namespace.id) return next(req);
    return next(req.clone({ setHeaders: { 'X-Namespace': namespace.id } }));
};

export function sanitizeMessage(message: string): string {
    try {
        const decoded: unknown = JSON.parse(message);
        return typeof decoded === 'string' ? decoded : message;
    } catch {
        return message;
    }
}

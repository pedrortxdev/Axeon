// frontend/lib/utils.ts
export const formatBytes = (bytes?: number) => {
    if (bytes === undefined || bytes === 0) return { value: 0, unit: 'MB' };
    const mb = bytes / 1024;
    if (mb < 1024) return { value: mb.toFixed(0), unit: 'MB' };
    return { value: (mb / 1024).toFixed(2), unit: 'GB' };
};

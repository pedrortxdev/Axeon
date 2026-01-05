import { useState, useEffect } from 'react';

export interface InstanceMetrics {
    cpuUsageSeconds: number;
    memoryUsageBytes: number;
    netRxBytes: number;
    netTxBytes: number;
    diskAllocatedBytes: number;
}

export const useInstanceControl = (instanceId: string, initialStatus: string) => {
    const [status, setStatus] = useState(initialStatus);
    const [metrics, setMetrics] = useState<InstanceMetrics>({
        cpuUsageSeconds: 0,
        memoryUsageBytes: 0,
        netRxBytes: 0,
        netTxBytes: 0,
        diskAllocatedBytes: 0
    });
    const [loadingAction, setLoadingAction] = useState(false);

    // Polling de Métricas
    useEffect(() => {
        if (status !== 'RUNNING') return;

        const fetchMetrics = async () => {
            try {
                const token = localStorage.getItem('axion_token');
                const protocol = window.location.protocol;
                const host = window.location.hostname;

                const res = await fetch(`${protocol}//${host}:8500/api/v1/instances/${instanceId}/metrics`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (res.ok) {
                    const data = await res.json();
                    setMetrics({
                        cpuUsageSeconds: (data.cpu_usage_us || 0) / 1_000_000,
                        memoryUsageBytes: data.memory_used_bytes || 0,
                        netRxBytes: data.net_rx_bytes || 0,
                        netTxBytes: data.net_tx_bytes || 0,
                        diskAllocatedBytes: data.disk_allocated_bytes || 0
                    });
                }
            } catch (e) {
                console.error("Erro ao buscar métricas", e);
            }
        };

        fetchMetrics();
        const interval = setInterval(fetchMetrics, 10000); // 10 segundos

        return () => clearInterval(interval);
    }, [instanceId, status]);

    // Power Actions
    const handlePowerAction = async (action: 'start' | 'stop' | 'reboot') => {
        setLoadingAction(true);
        try {
            const token = localStorage.getItem('axion_token');
            const protocol = window.location.protocol;
            const host = window.location.hostname;

            const res = await fetch(`${protocol}//${host}:8500/api/v1/instances/${instanceId}/action`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ action })
            });

            if (res.ok) {
                // Optimistic UI update
                if (action === 'start') setStatus('RUNNING');
                if (action === 'stop') setStatus('STOPPED');
                if (action === 'reboot') setStatus('RUNNING');
            } else {
                console.error(`Action ${action} failed:`, await res.text());
            }
        } catch (error) {
            console.error(`Falha ao executar ${action}:`, error);
        } finally {
            setLoadingAction(false);
        }
    };

    return { status, setStatus, metrics, handlePowerAction, loadingAction };
};

// Helper para formatar Bytes
export const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// Helper para formatar segundos de CPU
export const formatCpuTime = (seconds: number) => {
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
    return `${(seconds / 3600).toFixed(1)}h`;
};

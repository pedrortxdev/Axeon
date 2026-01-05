import React, { useEffect, useState } from 'react';
import { Globe, Lock } from 'lucide-react';
import { toast } from 'sonner';

interface Network {
    id: string;
    name: string;
    cidr: string;
    is_public: boolean;
    usage_percent: number;
}

interface Props {
    selectedId: string;
    onChange: (id: string) => void;
}

export const NetworkSelector: React.FC<Props> = ({ selectedId, onChange }) => {
    const [networks, setNetworks] = useState<Network[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchNetworks = async () => {
            try {
                const protocol = window.location.protocol;
                const host = window.location.hostname;
                const port = '8500';
                const token = localStorage.getItem('axion_token');
                if (!token) return;

                const res = await fetch(`${protocol}//${host}:${port}/api/v1/networks`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (res.ok) {
                    const data = await res.json();
                    setNetworks(data);
                    // Auto-select first if none selected
                    if (!selectedId && data.length > 0) {
                        onChange(data[0].id);
                    }
                }
            } catch (err) {
                console.error(err);
                toast.error("Failed to load networks");
            } finally {
                setLoading(false);
            }
        };

        fetchNetworks();
    }, [selectedId, onChange]);

    if (loading) return <div className="text-zinc-500 text-sm animate-pulse">Loading networks...</div>;

    return (
        <div className="space-y-3">
            <label className="block text-sm font-medium text-zinc-400">
                Network Location (IP Pool)
            </label>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                {networks.map((net) => (
                    <div
                        key={net.id}
                        onClick={() => onChange(net.id)}
                        className={`
              relative flex cursor-pointer rounded-lg border p-4 shadow-sm transition-all
              ${selectedId === net.id
                                ? 'border-indigo-500 ring-1 ring-indigo-500 bg-indigo-500/10'
                                : 'border-zinc-700 bg-zinc-900 hover:border-zinc-500'}
            `}
                    >
                        <div className="flex w-full items-center justify-between">
                            <div className="flex items-center">
                                <div className="text-sm">
                                    <p className={`font-medium ${selectedId === net.id ? 'text-indigo-300' : 'text-zinc-300'}`}>
                                        {net.name}
                                    </p>
                                    <p className="text-zinc-500 text-xs font-mono mt-1 bg-zinc-950 px-1.5 py-0.5 rounded inline-block">
                                        {net.cidr}
                                    </p>
                                </div>
                            </div>

                            <div className={`h-5 w-5 ${net.is_public ? 'text-purple-400' : 'text-emerald-400'}`}>
                                {net.is_public ? <Globe size={20} /> : <Lock size={20} />}
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            <p className="text-[10px] text-zinc-500 mt-2 flex items-center gap-1">
                * IP will be auto-allocated from the selected pool.
            </p>
        </div>
    );
};

'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Toaster, toast } from 'sonner';
import { ArrowLeft, Server, ShieldCheck, Globe, Trash } from 'lucide-react';

interface IpLease {
    ip_address: string;
    instance_name?: string;
    allocated_at?: string;
    status: string;
}

interface NetworkDetails {
    id: string;
    name: string;
    cidr: string;
    gateway: string;
    is_public: boolean;
    stats: {
        total_ips: number;
        used_ips: number;
        usage_percent: number;
    };
    leases: IpLease[];
}

export default function NetworkDetailsPage() {
    const params = useParams();
    const router = useRouter();
    const [network, setNetwork] = useState<NetworkDetails | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchNetwork = useCallback(async () => {
        try {
            const protocol = window.location.protocol;
            const host = window.location.hostname;
            const port = '8500';
            const token = localStorage.getItem('axion_token');
            if (!token) return;

            const res = await fetch(`${protocol}//${host}:${port}/api/v1/networks/${params.id}`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            if (res.ok) {
                const data = await res.json();
                setNetwork(data);
            } else {
                toast.error("Failed to load network details");
            }
        } catch (err) {
            console.error(err);
            toast.error("Network error");
        } finally {
            setLoading(false);
        }
    }, [params.id]);

    useEffect(() => {
        fetchNetwork();
    }, [fetchNetwork]);

    if (loading) {
        return <div className="p-8 text-zinc-500">Loading details...</div>;
    }

    if (!network) {
        return <div className="p-8 text-zinc-500">Network not found</div>;
    }

    return (
        <div className="p-6 transition-all fade-in">
            <Toaster position="top-right" theme="dark" />

            {/* Header */}
            <button
                onClick={() => router.back()}
                className="flex items-center gap-2 text-zinc-400 hover:text-white mb-6 transition-colors"
            >
                <ArrowLeft size={16} /> Back to Networks
            </button>

            <div className="flex justify-between items-end mb-8 border-b border-zinc-800 pb-6">
                <div>
                    <h1 className="text-3xl font-bold text-zinc-100 flex items-center gap-3">
                        {network.name}
                        <span className={`text-xs px-2.5 py-1 rounded-full border ${network.is_public ? 'bg-purple-500/10 text-purple-400 border-purple-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'}`}>
                            {network.is_public ? 'PUBLIC' : 'PRIVATE'}
                        </span>
                    </h1>
                    <p className="text-zinc-500 font-mono text-lg mt-1">{network.cidr}</p>
                </div>
                <div className="text-right">
                    <div className="text-2xl font-bold text-indigo-400">
                        {network.stats.used_ips} <span className="text-zinc-600 text-lg font-normal">/ {network.stats.total_ips} Alocated</span>
                    </div>
                    <p className="text-xs text-zinc-500 mt-1">Gateway: {network.gateway}</p>
                </div>
            </div>

            {/* Leases Table */}
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden shadow-lg backdrop-blur-sm">
                <div className="px-6 py-4 border-b border-zinc-800 bg-zinc-900 flex justify-between items-center">
                    <h3 className="font-semibold text-zinc-300">Allocated IPs</h3>
                    {network.leases.length === 0 && (
                        <span className="text-xs bg-zinc-800 text-zinc-500 px-2 py-1 rounded">Empty Pool</span>
                    )}
                </div>
                <table className="w-full text-left max-h-[600px] overflow-y-auto">
                    <thead className="bg-zinc-950/50 text-zinc-500 text-xs uppercase tracking-wider">
                        <tr>
                            <th className="px-6 py-3 font-medium">IP Address</th>
                            <th className="px-6 py-3 font-medium">Status</th>
                            <th className="px-6 py-3 font-medium">Associated VM</th>
                            <th className="px-6 py-3 font-medium">Allocated At</th>
                            <th className="px-6 py-3 font-medium text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800 text-zinc-300">
                        {network.leases.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="px-6 py-12 text-center text-zinc-500 italic">
                                    No active leases found.
                                </td>
                            </tr>
                        ) : network.leases.map((lease) => (
                            <tr key={lease.ip_address} className="hover:bg-zinc-800/30 transition-colors">
                                <td className="px-6 py-4 font-mono text-indigo-300 font-bold">
                                    {lease.ip_address}
                                </td>
                                <td className="px-6 py-4">
                                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-medium ${lease.status === 'allocated' ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-500'}`}>
                                        {lease.status === 'allocated' ? <ShieldCheck size={12} /> : <ShieldCheck size={12} />}
                                        {lease.status}
                                    </span>
                                </td>
                                <td className="px-6 py-4">
                                    {lease.instance_name ? (
                                        <div
                                            onClick={() => router.push(`/instances/${lease.instance_name}`)}
                                            className="flex items-center gap-2 cursor-pointer hover:text-white group"
                                        >
                                            <Server size={14} className="text-zinc-500 group-hover:text-indigo-400" />
                                            <span className="underline decoration-zinc-700 group-hover:decoration-indigo-500 underline-offset-4">{lease.instance_name}</span>
                                        </div>
                                    ) : (
                                        <span className="text-zinc-600 italic">System / Reserved</span>
                                    )}
                                </td>
                                <td className="px-6 py-4 text-sm text-zinc-500">
                                    {lease.allocated_at ? new Date(lease.allocated_at).toLocaleString() : '-'}
                                </td>
                                <td className="px-6 py-4 text-right">
                                    {lease.status === 'allocated' && (
                                        <button className="text-red-500/50 hover:text-red-400 text-xs font-bold uppercase tracking-wide hover:bg-red-500/10 px-2 py-1 rounded transition-colors">
                                            Force Release
                                        </button>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

        </div>
    );
}

'use client';

import React, { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { toast, Toaster } from 'sonner';
import {
    Play,
    Square,
    RefreshCw,
    Terminal,
    Cpu,
    HardDrive,
    Network,
    Copy,
    ChevronLeft,
    Server
} from 'lucide-react';
import Link from 'next/link';
import { useInstanceControl } from '@/hooks/useInstanceControl';

export default function InstanceDetails() {
    const params = useParams();
    const router = useRouter();
    const name = params?.name as string;

    const [instance, setInstance] = useState<any>(null);
    const [loading, setLoading] = useState(true);

    // Use our control hook for actions
    const { status, handlePowerAction, loadingAction } = useInstanceControl(name, 'UNKNOWN');

    useEffect(() => {
        if (!name) return;

        // Initial fetch of static details
        const fetchInstance = async () => {
            try {
                const token = localStorage.getItem('axion_token');
                if (!token) {
                    router.push('/login');
                    return;
                }

                const response = await fetch(`/api/v1/instances/${name}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (response.ok) {
                    const data = await response.json();
                    // Map backend fields to frontend state
                    setInstance({
                        ...data,
                        // Parse memory from limits if available, e.g. "512MB" -> 512
                        memory_mib: parseMemory(data.limits?.['limits.memory'] || data.limits?.['memory'] || '512MB'),
                        // Disk comes as bytes in disk_limit, convert to GB
                        disk_size_gb: (data.disk_limit || 0) / (1024 * 1024 * 1024),
                        // Default gateway for MVP
                        guest_gateway: '172.16.0.1',
                        bandwidth_limit_mbps: 100 // Hardcoded for MVP free tier
                    });
                } else {
                    if (response.status === 404) toast.error("Instance not found");
                }
            } catch (err) {
                console.error("Failed to fetch instance", err);
                toast.error("Failed to load instance details");
            } finally {
                setLoading(false);
            }
        };

        fetchInstance();
    }, [name, router]);

    // Helper to parse memory string
    const parseMemory = (mem: string) => {
        if (!mem) return 512;
        const m = mem.toUpperCase();
        if (m.endsWith('MB')) return parseInt(m);
        if (m.endsWith('GB')) return parseInt(m) * 1024;
        return parseInt(m);
    }

    if (loading) return (
        <div className="flex items-center justify-center min-h-screen bg-black">
            <div className="flex items-center gap-2 text-zinc-400">
                <div className="w-4 h-4 border-2 border-zinc-600 border-t-zinc-200 rounded-full animate-spin"></div>
                Loading VM data...
            </div>
        </div>
    );

    if (!instance) return (
        <div className="min-h-screen bg-black flex items-center justify-center text-zinc-400">
            <div>
                <h2 className="text-xl font-bold mb-2">VM Not Found</h2>
                <Link href="/instances" className="text-indigo-400 hover:underline">Return to instances</Link>
            </div>
        </div>
    );

    // Use status from hook if available (for optimistic updates), else from DB
    const currentStatus = (status !== 'UNKNOWN' ? status : instance.status) || 'UNKNOWN';
    const isRunning = currentStatus === 'RUNNING';

    return (
        <div className="min-h-screen bg-black text-zinc-200 p-6 md:p-10 font-sans selection:bg-indigo-500/30">
            <Toaster position="top-right" theme="dark" />

            <div className="max-w-5xl mx-auto space-y-8">

                {/* Back Link */}
                <Link href="/instances" className="inline-flex items-center text-sm text-zinc-500 hover:text-zinc-300 transition-colors">
                    <ChevronLeft size={16} className="mr-1" /> Back to Instances
                </Link>

                {/* 1. Header */}
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center bg-zinc-900/50 p-8 rounded-2xl border border-zinc-800/50 backdrop-blur-sm">
                    <div className="mb-6 md:mb-0">
                        <h1 className="text-4xl font-bold text-white mb-3 tracking-tight">{instance.name}</h1>
                        <div className="flex items-center space-x-3">
                            <span className="relative flex h-3 w-3">
                                <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${isRunning ? 'bg-emerald-500' : 'bg-red-500'} ${!isRunning && 'hidden'}`}></span>
                                <span className={`relative inline-flex rounded-full h-3 w-3 ${isRunning ? 'bg-emerald-500' : 'bg-red-500'}`}></span>
                            </span>
                            <span className={`font-mono text-sm font-medium tracking-wider ${isRunning ? 'text-emerald-400' : 'text-red-400'}`}>
                                {currentStatus}
                            </span>
                            <span className="text-zinc-600">|</span>
                            <span className="text-xs text-zinc-500 font-mono uppercase">{instance.node || 'Local Node'}</span>
                        </div>
                    </div>

                    {/* Controls */}
                    <div className="flex items-center gap-3">
                        {isRunning && (
                            <button
                                onClick={() => toast.info("Web Console coming soon!")}
                                className="bg-zinc-800 hover:bg-zinc-700 text-zinc-200 px-4 py-2.5 rounded-xl border border-zinc-700/50 font-medium transition-all flex items-center gap-2 text-sm"
                            >
                                <Terminal size={16} /> Console
                            </button>
                        )}

                        {!isRunning ? (
                            <button
                                onClick={() => handlePowerAction('start')}
                                disabled={loadingAction}
                                className="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-2.5 rounded-xl font-bold shadow-lg shadow-emerald-900/20 hover:shadow-emerald-900/40 transition-all active:scale-95 disabled:opacity-50 flex items-center gap-2"
                            >
                                <Play size={18} fill="currentColor" /> Start
                            </button>
                        ) : (
                            <>
                                <button
                                    onClick={() => handlePowerAction('reboot')}
                                    disabled={loadingAction}
                                    className="bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 hover:text-blue-300 px-4 py-2.5 rounded-xl border border-blue-600/20 font-medium transition-all flex items-center gap-2"
                                >
                                    <RefreshCw size={18} /> Restart
                                </button>
                                <button
                                    onClick={() => handlePowerAction('stop')}
                                    disabled={loadingAction}
                                    className="bg-red-600/10 hover:bg-red-600/20 text-red-400 hover:text-red-300 px-4 py-2.5 rounded-xl border border-red-600/20 font-medium transition-all flex items-center gap-2"
                                >
                                    <Square size={18} fill="currentColor" /> Stop
                                </button>
                            </>
                        )}
                    </div>
                </div>

                {/* 2. Connectivity Area */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

                    {/* SSH Card */}
                    <div className="lg:col-span-2 bg-zinc-900/30 p-8 rounded-2xl border border-zinc-800/50 flex flex-col justify-center">
                        <div className="flex items-center mb-6 text-indigo-400">
                            <Terminal className="w-6 h-6 mr-3" />
                            <h2 className="text-xl font-bold tracking-tight text-zinc-100">Access via SSH</h2>
                        </div>

                        <div className="bg-black/50 p-5 rounded-xl border border-zinc-800 flex justify-between items-center group hover:border-zinc-700 transition-colors">
                            <code className="text-emerald-400 font-mono text-sm md:text-base">
                                ssh root@{instance.ipAddress || 'Checking...'}
                            </code>
                            <button
                                onClick={() => {
                                    navigator.clipboard.writeText(`ssh root@${instance.ipAddress}`);
                                    toast.success("SSH command copied to clipboard");
                                }}
                                className="text-zinc-500 hover:text-zinc-200 transition-colors p-2"
                            >
                                <Copy size={18} />
                            </button>
                        </div>
                        <p className="text-xs text-zinc-500 mt-4 flex items-center gap-2">
                            <span className="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
                            Internal IP address. Configure VPN or Port Forwarding for external access.
                        </p>
                    </div>

                    {/* Network Summary */}
                    <div className="bg-zinc-900/30 p-8 rounded-2xl border border-zinc-800/50">
                        <div className="flex items-center mb-6 text-purple-400">
                            <Network className="w-6 h-6 mr-3" />
                            <h2 className="text-xl font-bold tracking-tight text-zinc-100">Network</h2>
                        </div>
                        <div className="space-y-4">
                            <div className="flex justify-between items-center border-b border-zinc-800/50 pb-3">
                                <span className="text-zinc-500 text-sm font-medium">Internal IP</span>
                                <span className="text-zinc-200 font-mono text-sm">{instance.ipAddress || 'N/A'}</span>
                            </div>
                            <div className="flex justify-between items-center border-b border-zinc-800/50 pb-3">
                                <span className="text-zinc-500 text-sm font-medium">Gateway</span>
                                <span className="text-zinc-200 font-mono text-sm">{instance.guest_gateway}</span>
                            </div>
                            <div className="flex justify-between items-center">
                                <span className="text-zinc-500 text-sm font-medium">Speed Limit</span>
                                <span className="text-amber-400 font-mono text-sm font-bold">{instance.bandwidth_limit_mbps} Mbps</span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* 3. Allocated Resources Cards */}
                <div>
                    <h2 className="text-lg font-semibold text-zinc-400 mb-4 px-1">Allocated Resources <span className="text-xs font-normal text-zinc-600 bg-zinc-900 border border-zinc-800 rounded px-1.5 py-0.5 ml-2">FREE TIER</span></h2>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">

                        {/* vCPU */}
                        <div className="bg-zinc-900/30 p-6 rounded-2xl border border-zinc-800/50 flex items-center hover:bg-zinc-900/50 transition-colors">
                            <div className="bg-blue-500/10 p-4 rounded-xl mr-5 border border-blue-500/10">
                                <Cpu className="w-8 h-8 text-blue-400" strokeWidth={1.5} />
                            </div>
                            <div>
                                <p className="text-xs text-zinc-500 uppercase font-bold tracking-wider mb-1">Processor</p>
                                <p className="text-3xl font-bold text-zinc-100 tracking-tight">{instance.cpu_count || 1} <span className="text-base font-medium text-zinc-500">vCore</span></p>
                            </div>
                        </div>

                        {/* RAM */}
                        <div className="bg-zinc-900/30 p-6 rounded-2xl border border-zinc-800/50 flex items-center hover:bg-zinc-900/50 transition-colors">
                            <div className="bg-purple-500/10 p-4 rounded-xl mr-5 border border-purple-500/10">
                                <Server className="w-8 h-8 text-purple-400" strokeWidth={1.5} />
                            </div>
                            <div>
                                <p className="text-xs text-zinc-500 uppercase font-bold tracking-wider mb-1">Memory</p>
                                <p className="text-3xl font-bold text-zinc-100 tracking-tight">{instance.memory_mib} <span className="text-base font-medium text-zinc-500">MB</span></p>
                            </div>
                        </div>

                        {/* DISK */}
                        <div className="bg-zinc-900/30 p-6 rounded-2xl border border-zinc-800/50 flex items-center hover:bg-zinc-900/50 transition-colors">
                            <div className="bg-amber-500/10 p-4 rounded-xl mr-5 border border-amber-500/10">
                                <HardDrive className="w-8 h-8 text-amber-400" strokeWidth={1.5} />
                            </div>
                            <div>
                                <p className="text-xs text-zinc-500 uppercase font-bold tracking-wider mb-1">Storage</p>
                                <p className="text-3xl font-bold text-zinc-100 tracking-tight">{Math.round(instance.disk_size_gb)} <span className="text-base font-medium text-zinc-500">GB</span></p>
                            </div>
                        </div>

                    </div>
                </div>

                {/* Footer */}
                <div className="text-xs text-zinc-600 font-mono pt-8 border-t border-zinc-900 flex justify-between">
                    <span>ID: {instance.name}</span>
                    <span>Image: {instance.image}</span>
                </div>

            </div>
        </div>
    );
}

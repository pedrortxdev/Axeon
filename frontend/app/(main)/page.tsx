'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import ClusterStatus from '@/components/ClusterStatus';
import HostStatsCard from '@/components/HostStatsCard';
import { HostStats } from '@/types';

export default function OverviewPage() {
  const [stats, setStats] = useState<HostStats | null>(null);
  const storedToken = typeof window !== 'undefined' ? localStorage.getItem('axion_token') : null;
  const [token] = useState<string | null>(storedToken);
  const wsRef = useRef<WebSocket | null>(null);
  const router = useRouter();

  // --- Auth Check ---
  useEffect(() => {
    if (!token) {
        router.push('/login');
    }
  }, [router, token]);

  // --- WebSocket Logic ---
  useEffect(() => {
    if (!token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const wsUrl = `${protocol}//${host}:8500/ws/telemetry?token=${token}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const rawData = JSON.parse(event.data);
        if (rawData && rawData.type === 'host_telemetry') {
            setStats(rawData.data);
        }
      } catch (err) {
        console.error('WS parsing error:', err);
      }
    };

    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, [token]);

  if (!token) return null;

  return (
    <div className="max-w-7xl mx-auto">
      <header className="mb-8">
        <h1 className="text-2xl font-bold text-zinc-100">Dashboard Overview</h1>
        <p className="text-zinc-500">Monitor your infrastructure at a glance</p>
      </header>

      <div className="space-y-8">
        <HostStatsCard data={stats} />
        <ClusterStatus token={token} />
      </div>
    </div>
  );
}
'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Sidebar from '@/components/Sidebar';

export default function MainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('axion_token');
    if (!token) {
      router.push('/login');
    } else {
      setIsMounted(true);

      // Fetch branding
      const protocol = window.location.protocol;
      const hostname = window.location.hostname;
      fetch(`${protocol}//${hostname}:8500/branding/settings`, {
        headers: { 'Authorization': `Bearer ${token}` }
      })
        .then(res => res.json())
        .then(data => {
          if (data.primary_color) {
            // Apply color override
            const style = document.createElement('style');
            style.innerHTML = `
              :root { --primary-color: ${data.primary_color}; }
              .bg-blue-600, .hover\\:bg-blue-600:hover { background-color: ${data.primary_color} !important; }
              .text-blue-600 { color: ${data.primary_color} !important; }
              .border-blue-600 { border-color: ${data.primary_color} !important; }
            `;
            document.head.appendChild(style);
          }
        })
        .catch(console.error);
    }
  }, [router]);


  if (!isMounted) {
    return null;
  }

  return (
    <div className="flex h-screen bg-zinc-950 text-white">
      <Sidebar />
      <main className="flex-1 overflow-auto p-8">
        {children}
      </main>
    </div>
  );
}
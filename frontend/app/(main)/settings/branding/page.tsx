'use client';

import React, { useState, useEffect } from 'react';
import { toast } from 'sonner';

export default function BrandingSettingsPage() {
    const [logoPreview, setLogoPreview] = useState<string | null>(null);
    const [primaryColor, setPrimaryColor] = useState('#3B82F6');
    const [hidePoweredBy, setHidePoweredBy] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    // Helper to get API URL
    const getApiUrl = () => {
        if (typeof window === 'undefined') return '';
        const protocol = window.location.protocol;
        const hostname = window.location.hostname;
        return `${protocol}//${hostname}:8500`;
    };

    const token = typeof window !== 'undefined' ? localStorage.getItem('axion_token') : null;
    const apiUrl = getApiUrl();

    useEffect(() => {
        if (!token) return;
        fetch(`${apiUrl}/branding/settings`, {
            headers: { 'Authorization': `Bearer ${token}` }
        })
            .then(res => res.json())
            .then(data => {
                if (data.logo_url) setLogoPreview(data.logo_url);
                if (data.primary_color) setPrimaryColor(data.primary_color);
                if (data.hide_powered_by !== undefined) setHidePoweredBy(data.hide_powered_by);
                setIsLoading(false);
            })
            .catch(() => {
                // Silence error if just not set or network init
                setIsLoading(false);
            });
    }, [token, apiUrl]);

    const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (!e.target.files || !e.target.files[0]) return;
        const file = e.target.files[0];

        const formData = new FormData();
        formData.append('logo', file);

        try {
            const res = await fetch(`${apiUrl}/branding/upload-logo`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` },
                body: formData
            });
            if (res.ok) {
                const data = await res.json();
                setLogoPreview(data.url);
                toast.success("Logo uploaded");
                // Reload to apply immediately if logic is in layout
                setTimeout(() => window.location.reload(), 500);
            } else {
                toast.error("Upload failed");
            }
        } catch (err) {
            toast.error("Network error");
        }
    };

    const handleSave = async () => {
        try {
            const res = await fetch(`${apiUrl}/branding/settings`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ primary_color: primaryColor, hide_powered_by: hidePoweredBy })
            });
            if (res.ok) {
                toast.success("Settings saved");
                setTimeout(() => window.location.reload(), 500);
            } else {
                toast.error("Save failed");
            }
        } catch (err) {
            toast.error("Network error");
        }
    };

    if (isLoading) return <div className="p-8 text-zinc-400">Loading settings...</div>;

    const logoSrc = logoPreview ? (logoPreview.startsWith('http') ? logoPreview : `${apiUrl}${logoPreview}`) : null;

    return (
        <div className="max-w-2xl space-y-8">
            <div className="space-y-2">
                <h1 className="text-2xl font-bold text-white">Branding & White Label</h1>
                <p className="text-zinc-400">Customize the look and feel of your Axion instance.</p>
            </div>

            <div className="space-y-6">
                {/* Logo Section */}
                <div className="p-6 rounded-xl bg-zinc-900/50 border border-zinc-800 space-y-4">
                    <h2 className="text-lg font-semibold text-white">Custom Logo</h2>
                    <div className="flex items-start gap-6">
                        <div className="w-24 h-24 bg-zinc-950 rounded-lg flex items-center justify-center border border-zinc-800 overflow-hidden">
                            {logoSrc ? (
                                <img src={logoSrc} alt="Logo Preview" className="max-w-full max-h-full object-contain" />
                            ) : (
                                <span className="text-xs text-zinc-600">No Logo</span>
                            )}
                        </div>
                        <div className="space-y-2 flex-1">
                            <label className="block text-sm text-zinc-400">Upload a new logo (PNG, JPG, SVG - Max 2MB)</label>
                            <input
                                type="file"
                                onChange={handleLogoUpload}
                                accept="image/*"
                                className="block w-full text-sm text-zinc-400
                                file:mr-4 file:py-2 file:px-4
                                file:rounded-full file:border-0
                                file:text-sm file:font-semibold
                                file:bg-[#3B82F6] file:text-white
                                hover:file:bg-blue-600
                                cursor-pointer"
                            />
                        </div>
                    </div>
                </div>

                {/* Color Section */}
                <div className="p-6 rounded-xl bg-zinc-900/50 border border-zinc-800 space-y-4">
                    <h2 className="text-lg font-semibold text-white">Primary Color</h2>
                    <div className="flex items-center gap-4">
                        <input
                            type="color"
                            value={primaryColor}
                            onChange={e => setPrimaryColor(e.target.value)}
                            className="h-12 w-24 bg-transparent border-none cursor-pointer rounded"
                        />
                        <div className="space-y-1">
                            <p className="text-white font-mono">{primaryColor}</p>
                            <p className="text-xs text-zinc-500">This color will be applied to buttons and active states.</p>
                        </div>
                    </div>
                </div>

                {/* Footer Section */}
                <div className="p-6 rounded-xl bg-zinc-900/50 border border-zinc-800 space-y-4">
                    <h2 className="text-lg font-semibold text-white">Footer</h2>
                    <div className="flex items-center gap-3">
                        <input
                            type="checkbox"
                            checked={hidePoweredBy}
                            onChange={e => setHidePoweredBy(e.target.checked)}
                            id="hidePoweredBy"
                            className="h-5 w-5 rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-blue-600 focus:ring-offset-zinc-900"
                        />
                        <label htmlFor="hidePoweredBy" className="text-zinc-300">Hide "Powered by Axion" in footer</label>
                    </div>
                </div>

                <div className="pt-4">
                    <button
                        onClick={handleSave}
                        className="bg-[#3B82F6] hover:bg-blue-600 text-white px-6 py-2.5 rounded-lg font-medium transition-colors"
                        style={{ backgroundColor: primaryColor }}
                    >
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
    );
}

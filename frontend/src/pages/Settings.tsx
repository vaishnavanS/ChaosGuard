import { useState } from 'react';
import { api, getBackendURL } from '../services/api';
import { 
  Database, 
  Sun,
  Moon,
  CheckCircle,
  XCircle,
  Clock
} from 'lucide-react';

export default function Settings() {
  const [backendUrl, setBackendUrl] = useState(() => getBackendURL());
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('chaosguard_theme') as 'light' | 'dark') || 'dark';
  });
  const [refreshInterval, setRefreshInterval] = useState(() => {
    return parseInt(localStorage.getItem('chaosguard_refresh_interval') || '5000', 10);
  });

  // Test Connection status
  const [connTest, setConnTest] = useState<'idle' | 'testing' | 'success' | 'failed'>('idle');

  const handleSaveSettings = (e: React.FormEvent) => {
    e.preventDefault();
    localStorage.setItem('chaosguard_backend_url', backendUrl);
    localStorage.setItem('chaosguard_theme', theme);
    localStorage.setItem('chaosguard_refresh_interval', refreshInterval.toString());
    
    // Trigger theme update on html root
    const root = window.document.documentElement;
    if (theme === 'dark') {
      root.classList.add('dark');
      root.style.backgroundColor = '#0b0f19';
    } else {
      root.classList.remove('dark');
      root.style.backgroundColor = '#f9fafb';
    }

    alert('Settings successfully updated and saved locally.');
  };

  const handleTestConnection = async () => {
    setConnTest('testing');
    try {
      // Temporary override URL to test target endpoint directly
      localStorage.setItem('chaosguard_backend_url', backendUrl);
      const res = await api.getHealth();
      if (res && res.success) {
        setConnTest('success');
      } else {
        setConnTest('failed');
      }
    } catch {
      setConnTest('failed');
    }
  };

  return (
    <div className="space-y-8 max-w-2xl">
      
      <form onSubmit={handleSaveSettings} className="space-y-6">
        
        {/* API connection preferences */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 space-y-4">
          <div className="flex items-center gap-2 text-violet-400">
            <Database className="h-5 w-5" />
            <h2 className="text-base font-semibold tracking-tight">API Server Address</h2>
          </div>
          
          <div className="space-y-2">
            <label className="text-xs font-bold uppercase tracking-wider text-slate-500">Backend Daemon Base URL</label>
            <div className="flex flex-col sm:flex-row gap-3 items-stretch">
              <input 
                type="url" 
                required
                value={backendUrl}
                onChange={(e) => setBackendUrl(e.target.value)}
                className="flex-1 bg-slate-900/60 border border-slate-800 rounded-lg px-3 py-2 text-sm text-gray-200 focus:outline-hidden focus:border-violet-500"
                placeholder="http://localhost:8080"
              />
              <button
                type="button"
                onClick={handleTestConnection}
                disabled={connTest === 'testing'}
                className="px-4 py-2 border border-slate-800 rounded-lg text-xs font-bold uppercase tracking-wider hover:bg-slate-800/40 cursor-pointer disabled:opacity-50 flex items-center justify-center gap-2 select-none"
              >
                {connTest === 'testing' && 'Testing...'}
                {connTest === 'idle' && 'Test API'}
                {connTest === 'success' && (
                  <>
                    <CheckCircle className="h-4 w-4 text-emerald-400" />
                    <span className="text-emerald-400">Connected</span>
                  </>
                )}
                {connTest === 'failed' && (
                  <>
                    <XCircle className="h-4 w-4 text-rose-400" />
                    <span className="text-rose-400">Failed</span>
                  </>
                )}
              </button>
            </div>
            <p className="text-[10px] text-slate-500 leading-relaxed mt-1">
              ChaosGuard daemon REST service address configured in `chaosguard.yaml` (default is port 8080).
            </p>
          </div>
        </div>

        {/* Dashboard theme preferences */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 space-y-4">
          <div className="flex items-center gap-2 text-amber-500">
            <Sun className="h-5 w-5" />
            <h2 className="text-base font-semibold tracking-tight">Display Configuration</h2>
          </div>

          <div className="space-y-4">
            
            {/* Theme selection */}
            <div className="flex flex-col gap-2">
              <label className="text-xs font-bold uppercase tracking-wider text-slate-500">Interface Theme</label>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => setTheme('dark')}
                  className={`py-3 px-4 rounded-lg border text-sm font-semibold flex items-center justify-center gap-2 cursor-pointer transition-all duration-150 ${
                    theme === 'dark'
                      ? 'bg-violet-600/10 text-violet-400 border-violet-500/30'
                      : 'border-slate-800 text-slate-400 hover:bg-slate-850'
                  }`}
                >
                  <Moon className="h-4 w-4" />
                  <span>Dark Mode (Recommended)</span>
                </button>
                <button
                  type="button"
                  onClick={() => setTheme('light')}
                  className={`py-3 px-4 rounded-lg border text-sm font-semibold flex items-center justify-center gap-2 cursor-pointer transition-all duration-150 ${
                    theme === 'light'
                      ? 'bg-violet-600/10 text-violet-400 border-violet-500/30'
                      : 'border-slate-800 text-slate-400 hover:bg-slate-850'
                  }`}
                >
                  <Sun className="h-4 w-4" />
                  <span>Light Mode</span>
                </button>
              </div>
            </div>

            {/* Refresh Interval selection */}
            <div className="flex flex-col gap-2">
              <label className="text-xs font-bold uppercase tracking-wider text-slate-500 flex items-center gap-1">
                <Clock className="h-3.5 w-3.5" /> Auto Refresh Period
              </label>
              <select
                value={refreshInterval}
                onChange={(e) => setRefreshInterval(parseInt(e.target.value, 10))}
                className="bg-slate-900/60 border border-slate-800 rounded-lg px-3 py-2 text-sm text-gray-200 focus:outline-hidden"
              >
                <option value="2500">2.5 seconds (High frequency)</option>
                <option value="5000">5 seconds (Recommended)</option>
                <option value="10000">10 seconds (Standard)</option>
                <option value="30000">30 seconds (Low overhead)</option>
              </select>
              <p className="text-[10px] text-slate-500 mt-1">
                Specifies how frequently the client dashboard polls new state parameters from the REST endpoints.
              </p>
            </div>

          </div>
        </div>

        {/* Form controls submit */}
        <div className="flex justify-end gap-3 shrink-0">
          <button
            type="submit"
            className="px-6 py-2.5 bg-violet-600 hover:bg-violet-700 text-white rounded-lg text-xs font-bold uppercase tracking-wider cursor-pointer"
          >
            Apply Configurations
          </button>
        </div>

      </form>

    </div>
  );
}

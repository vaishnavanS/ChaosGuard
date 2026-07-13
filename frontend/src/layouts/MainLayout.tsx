import { useEffect, useState, useRef } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import type { Experiment } from '../types';
import { AnimatePresence, motion } from 'framer-motion';
import { 
  LayoutDashboard, 
  Layers, 
  History, 
  Activity, 
  Terminal, 
  Settings, 
  Cpu,
  CloudLightning,
  Sun,
  Moon,
  WifiOff,
  Lightbulb,
  Bell,
  X,
  ShieldCheck,
  CheckCircle,
  AlertTriangle,
  Flame
} from 'lucide-react';

interface Toast {
  id: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
}

export default function MainLayout() {
  const location = useLocation();
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('chaosguard_theme') as 'light' | 'dark') || 'dark';
  });

  const [toasts, setToasts] = useState<Toast[]>([]);
  const prevExperimentsRef = useRef<Experiment[]>([]);
  const prevOnlineRef = useRef<boolean>(true);

  // Apply theme class to HTML root
  useEffect(() => {
    const root = window.document.documentElement;
    if (theme === 'dark') {
      root.classList.add('dark');
      root.style.backgroundColor = '#0b0f19';
    } else {
      root.classList.remove('dark');
      root.style.backgroundColor = '#f9fafb';
    }
    localStorage.setItem('chaosguard_theme', theme);
  }, [theme]);

  // Query health periodically to determine daemon status
  const { data: healthRes, isError } = useQuery({
    queryKey: ['healthState'],
    queryFn: () => api.getHealth(),
    refetchInterval: 5000,
    retry: 1,
  });

  const isOnline = !isError && healthRes?.success;

  // Query experiments to detect new attacks or recoveries for toasts
  const { data: experimentsRes } = useQuery({
    queryKey: ['experimentsToasts'],
    queryFn: () => api.getExperiments(),
    refetchInterval: 4000,
    enabled: isOnline,
  });

  // Toast Trigger Engine
  useEffect(() => {
    // 1. Online state transition toasts
    if (isOnline !== prevOnlineRef.current) {
      if (isOnline) {
        addToast('ChaosGuard daemon connection restored.', 'success');
      } else {
        addToast('Docker daemon or ChaosGuard REST API unreachable!', 'error');
      }
      prevOnlineRef.current = !!isOnline;
    }

    if (!isOnline || !experimentsRes?.data) return;

    const currentExps = experimentsRes.data;
    const prevExps = prevExperimentsRef.current;

    if (prevExps.length > 0) {
      // Find new experiments
      currentExps.forEach(curr => {
        const foundPrev = prevExps.find(p => p.id === curr.id);
        if (!foundPrev) {
          addToast(`Chaos Experiment Started: ${curr.attack_type.toUpperCase()} injected against ${curr.container_name}`, 'warning');
        } else if (curr.status !== foundPrev.status) {
          // Status change toast
          if (curr.status === 'completed' || curr.status === 'recovered') {
            addToast(`Recovery Completed: ${curr.container_name} restored successfully.`, 'success');
          } else if (curr.status === 'failed') {
            addToast(`Container Failed: ${curr.container_name} crashed during stress.`, 'error');
          }
        }
      });
    }

    prevExperimentsRef.current = currentExps;
  }, [isOnline, experimentsRes]);

  const addToast = (message: string, type: Toast['type']) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts(prev => [...prev, { id, message, type }]);
    setTimeout(() => {
      removeToast(id);
    }, 6000);
  };

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  const toggleTheme = () => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light');
  };

  // Grouped Navigation per specifications
  const navGroups = [
    {
      title: 'Overview',
      items: [
        { to: '/', label: 'Dashboard', icon: LayoutDashboard },
      ]
    },
    {
      title: 'Operations',
      items: [
        { to: '/containers', label: 'Containers', icon: Layers },
      ]
    },
    {
      title: 'Monitoring',
      items: [
        { to: '/metrics', label: 'Metrics', icon: Activity },
        { to: '/logs', label: 'Live Logs', icon: Terminal },
      ]
    },
    {
      title: 'Analysis',
      items: [
        { to: '/experiments', label: 'Experiments', icon: History },
        { to: '/recommendations', label: 'Recommendations', icon: Lightbulb },
      ]
    },
    {
      title: 'Administration',
      items: [
        { to: '/runtime', label: 'Runtime Status', icon: Cpu },
        { to: '/settings', label: 'Settings', icon: Settings },
      ]
    }
  ];

  const currentPathLabel = () => {
    for (const group of navGroups) {
      const match = group.items.find(i => i.to === location.pathname);
      if (match) return match.label;
    }
    return 'ChaosGuard';
  };

  return (
    <div className={`min-h-screen flex flex-col md:flex-row theme-transition ${theme === 'dark' ? 'bg-[#0b0f19] text-gray-100' : 'bg-gray-50 text-gray-800'}`}>
      
      {/* Offline Alert Banner */}
      {!isOnline && (
        <div className="absolute top-0 left-0 right-0 z-50 bg-rose-600/90 text-white py-2 px-4 text-center text-xs font-bold flex items-center justify-center gap-2 backdrop-blur-md">
          <WifiOff className="h-4 w-4 animate-pulse" />
          ChaosGuard Daemon is offline. Please start it using 'chaosguard start' or update the API address in settings.
        </div>
      )}

      {/* Persistent Left Sidebar */}
      <aside className={`hidden md:flex flex-col w-64 border-r ${theme === 'dark' ? 'bg-[#0c101b] border-slate-900' : 'bg-white border-gray-200'} shrink-0`}>
        <div className="h-16 flex items-center gap-3 px-6 border-b border-inherit">
          <CloudLightning className="h-6 w-6 text-violet-500 animate-pulse-status" />
          <span className="font-extrabold text-sm tracking-wider bg-gradient-to-r from-violet-400 to-indigo-500 bg-clip-text text-transparent">
            CHAOSGUARD
          </span>
        </div>

        <nav className="flex-1 px-4 py-6 space-y-6 overflow-y-auto">
          {navGroups.map((group, idx) => (
            <div key={idx} className="space-y-1.5">
              <span className="text-[10px] font-bold uppercase tracking-widest text-slate-500 px-3">
                {group.title}
              </span>
              <div className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  return (
                    <NavLink
                      key={item.to}
                      to={item.to}
                      className={({ isActive }) =>
                        `flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-semibold uppercase tracking-wider transition-all duration-200 border ${
                          isActive
                            ? 'bg-violet-600/10 text-violet-400 border-violet-500/20 shadow-sm'
                            : theme === 'dark' 
                              ? 'text-gray-400 border-transparent hover:bg-slate-800/40 hover:text-gray-100' 
                              : 'text-gray-600 border-transparent hover:bg-gray-100 hover:text-gray-900'
                        }`
                      }
                    >
                      <Icon className="h-4 w-4 shrink-0" />
                      <span className="truncate">{item.label}</span>
                    </NavLink>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        {/* Footer Info bar */}
        <div className={`p-4 border-t border-inherit text-[10px] uppercase font-bold tracking-wider ${theme === 'dark' ? 'text-slate-600 bg-[#070b13]/40' : 'text-gray-400 bg-gray-50'}`}>
          ChaosGuard Core v0.3.1
        </div>
      </aside>

      {/* Main Workspace Frame */}
      <div className="flex-1 flex flex-col min-w-0 overflow-y-auto">
        <header className={`h-16 flex items-center justify-between px-6 md:px-8 border-b ${theme === 'dark' ? 'bg-[#0c101b]/60 border-slate-900' : 'bg-white border-gray-200'} shrink-0 pt-2`}>
          <div>
            <h1 className="text-xs uppercase font-bold tracking-widest text-slate-500 my-0 py-0">
              {currentPathLabel()}
            </h1>
          </div>
          
          <div className="flex items-center gap-4">
            {/* Health / Connection State Indicators */}
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
              isOnline 
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/10' 
                : 'bg-rose-500/10 text-rose-400 border border-rose-500/10'
            }`}>
              <span className={`h-1.5 w-1.5 rounded-full ${isOnline ? 'bg-emerald-400' : 'bg-rose-400 animate-pulse'}`}></span>
              {isOnline ? 'Daemon Online' : 'Daemon Offline'}
            </div>

            {/* Dark Mode Switcher */}
            <button 
              onClick={toggleTheme} 
              className={`p-2 rounded-lg border ${theme === 'dark' ? 'border-slate-850 text-amber-400 hover:bg-slate-800' : 'border-gray-200 text-slate-600 hover:bg-gray-100'}`}
              title="Toggle Theme"
            >
              {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </button>
          </div>
        </header>

        <main className={`flex-1 p-6 md:p-8 ${!isOnline ? 'pt-16' : ''}`}>
          <Outlet />
        </main>
      </div>

      {/* Floating Toast Notification Containers */}
      <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-3 w-80 pointer-events-none">
        <AnimatePresence>
          {toasts.map(toast => (
            <motion.div
              key={toast.id}
              layout
              initial={{ opacity: 0, y: 30, scale: 0.9 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, scale: 0.85, transition: { duration: 0.15 } }}
              className={`p-4 rounded-xl border pointer-events-auto flex items-start justify-between gap-3 shadow-xl backdrop-blur-md ${
                toast.type === 'success' 
                  ? 'bg-emerald-950/80 border-emerald-500/20 text-emerald-300' 
                  : toast.type === 'warning'
                    ? 'bg-amber-950/80 border-amber-500/20 text-amber-300'
                    : toast.type === 'error'
                      ? 'bg-rose-950/80 border-rose-500/20 text-rose-300'
                      : 'bg-slate-900/80 border-slate-700/20 text-slate-300'
              }`}
            >
              <div className="flex gap-2">
                {toast.type === 'success' && <CheckCircle className="h-4.5 w-4.5 text-emerald-450 shrink-0 mt-0.5" />}
                {toast.type === 'error' && <Flame className="h-4.5 w-4.5 text-rose-450 shrink-0 mt-0.5 animate-pulse" />}
                {toast.type === 'warning' && <AlertTriangle className="h-4.5 w-4.5 text-amber-450 shrink-0 mt-0.5" />}
                {toast.type === 'info' && <Bell className="h-4.5 w-4.5 text-sky-450 shrink-0 mt-0.5" />}
                <p className="text-xs font-semibold leading-relaxed">{toast.message}</p>
              </div>
              <button 
                onClick={() => removeToast(toast.id)}
                className="p-0.5 rounded hover:bg-black/10 text-inherit shrink-0 cursor-pointer"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

    </div>
  );
}

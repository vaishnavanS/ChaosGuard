import { useEffect, useState } from 'react';
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
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
  Wifi,
  WifiOff,
  Menu,
  X
} from 'lucide-react';

export default function MainLayout() {
  const location = useLocation();
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('chaosguard_theme') as 'light' | 'dark') || 'dark';
  });
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  // Apply theme class to HTML element on change
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
  const { data: healthData, isError } = useQuery({
    queryKey: ['healthState'],
    queryFn: () => api.getHealth(),
    refetchInterval: 5000,
    retry: 1,
  });

  const isOnline = !isError && healthData?.success;

  const toggleTheme = () => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light');
  };

  const navItems = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard },
    { to: '/containers', label: 'Containers', icon: Layers },
    { to: '/experiments', label: 'Experiments', icon: History },
    { to: '/metrics', label: 'Metrics', icon: Activity },
    { to: '/runtime', label: 'Runtime', icon: Cpu },
    { to: '/logs', label: 'Live Logs', icon: Terminal },
    { to: '/settings', label: 'Settings', icon: Settings },
  ];

  return (
    <div className={`min-h-screen flex flex-col md:flex-row theme-transition ${theme === 'dark' ? 'bg-[#0b0f19] text-gray-100' : 'bg-gray-50 text-gray-800'}`}>
      
      {/* Offline Alert Banner */}
      {!isOnline && (
        <div className="absolute top-0 left-0 right-0 z-50 bg-rose-600/90 text-white py-2 px-4 text-center text-sm font-semibold flex items-center justify-center gap-2 backdrop-blur-md">
          <WifiOff className="h-4 w-4 animate-pulse" />
          ChaosGuard Daemon is offline. Please start it using 'chaosguard start' or update the API address in settings.
        </div>
      )}

      {/* Sidebar for Desktop */}
      <aside className={`hidden md:flex flex-col w-64 border-r ${theme === 'dark' ? 'bg-[#0f172a]/80 border-slate-800' : 'bg-white border-gray-200'} shrink-0`}>
        <div className="h-16 flex items-center gap-3 px-6 border-b border-inherit">
          <CloudLightning className="h-6 w-6 text-violet-500 animate-pulse-status" />
          <span className="font-bold text-lg tracking-wider bg-gradient-to-r from-violet-400 to-indigo-500 bg-clip-text text-transparent">
            CHAOSGUARD
          </span>
        </div>

        <nav className="flex-1 px-4 py-6 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                    isActive
                      ? 'bg-violet-600/15 text-violet-400 border border-violet-500/20 shadow-sm'
                      : theme === 'dark' 
                        ? 'text-gray-400 hover:bg-slate-800/50 hover:text-gray-100' 
                        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                  }`
                }
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            );
          })}
        </nav>

        {/* Footer info */}
        <div className={`p-4 border-t border-inherit text-xs ${theme === 'dark' ? 'text-slate-500' : 'text-gray-400'}`}>
          ChaosGuard Core v0.3.0
        </div>
      </aside>

      {/* Mobile Drawer Menu */}
      <div className="md:hidden flex items-center justify-between p-4 border-b border-inherit bg-slate-900/40 backdrop-blur-md">
        <div className="flex items-center gap-2">
          <CloudLightning className="h-5 w-5 text-violet-500" />
          <span className="font-bold text-sm tracking-wide text-white">CHAOSGUARD</span>
        </div>
        <button 
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          className="p-1.5 rounded-lg border border-slate-700 text-gray-300"
        >
          {isMobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {isMobileMenuOpen && (
        <div className="md:hidden absolute inset-0 z-40 bg-[#0b0f19]/95 backdrop-blur-lg flex flex-col p-6 space-y-6">
          <div className="flex justify-between items-center">
            <span className="font-bold text-lg text-violet-400">Navigation</span>
            <button onClick={() => setIsMobileMenuOpen(false)} className="p-1 rounded-full border border-slate-800">
              <X className="h-5 w-5" />
            </button>
          </div>
          <nav className="space-y-2">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.to}
                  to={item.to}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`flex items-center gap-3 px-4 py-3.5 rounded-lg text-base font-medium ${
                    location.pathname === item.to ? 'bg-violet-600/20 text-violet-400' : 'text-gray-400'
                  }`}
                >
                  <Icon className="h-5 w-5" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>
      )}

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-y-auto">
        <header className={`h-16 flex items-center justify-between px-6 md:px-8 border-b ${theme === 'dark' ? 'bg-[#0f172a]/20 border-slate-800' : 'bg-white border-gray-200'} shrink-0 pt-2`}>
          <div>
            <h1 className="text-xl font-semibold tracking-tight my-0 py-0">
              {navItems.find(i => i.to === location.pathname)?.label || 'ChaosGuard'}
            </h1>
          </div>
          
          <div className="flex items-center gap-4">
            {/* Online Status Dot */}
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold ${
              isOnline 
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' 
                : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
            }`}>
              {isOnline ? (
                <>
                  <Wifi className="h-3.5 w-3.5" />
                  <span>ONLINE</span>
                </>
              ) : (
                <>
                  <WifiOff className="h-3.5 w-3.5 animate-pulse" />
                  <span>OFFLINE</span>
                </>
              )}
            </div>

            {/* Dark Mode Switcher */}
            <button 
              onClick={toggleTheme} 
              className={`p-2 rounded-lg border ${theme === 'dark' ? 'border-slate-800 text-amber-400 hover:bg-slate-800' : 'border-gray-200 text-slate-600 hover:bg-gray-100'}`}
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

    </div>
  );
}

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import type { Container } from '../types';
import { 
  Play, 
  Pause, 
  Square, 
  RefreshCw, 
  Skull, 
  Cpu, 
  Database, 
  Clock, 
  AlertTriangle,
  Search,
  ShieldCheck,
  ShieldAlert,
  Loader2
} from 'lucide-react';

export default function Containers() {
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [filterMode, setFilterMode] = useState<'all' | 'monitored' | 'unmonitored'>('all');
  
  // Dialog state
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    containerId: string;
    containerName: string;
    attackType: string;
  } | null>(null);

  // Queries
  const { data: response, isLoading, isError, refetch } = useQuery({
    queryKey: ['containersList'],
    queryFn: () => api.getContainers(),
    refetchInterval: 3000,
  });

  const { data: experimentsRes } = useQuery({
    queryKey: ['experimentsListContainers'],
    queryFn: () => api.getExperiments(),
    refetchInterval: 3000,
  });

  // Mutations
  const injectAttack = useMutation({
    mutationFn: ({ id, type }: { id: string; type: string }) => 
      api.createExperiment(id, type, 15), // Default duration 15s
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['containersList'] });
      setConfirmDialog(null);
    },
    onError: (err: any) => {
      alert(`Failed to trigger chaos attack: ${err.message || 'unknown error'}`);
      setConfirmDialog(null);
    }
  });

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
        <Loader2 className="animate-spin h-8 w-8 text-violet-500" />
        <p className="text-sm text-gray-400">Querying active containers on Docker host...</p>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] text-center p-6 border border-slate-800 rounded-xl bg-slate-900/10">
        <AlertTriangle className="h-10 w-10 text-rose-500 mb-3" />
        <h3 className="text-lg font-semibold">Failed to load container data</h3>
        <p className="text-xs text-slate-500 mt-1 mb-4">The daemon may be offline or Docker Socket access is denied.</p>
        <button 
          onClick={() => refetch()}
          className="px-4 py-2 bg-violet-600 rounded-lg hover:bg-violet-700 text-sm font-semibold cursor-pointer"
        >
          Retry Connection
        </button>
      </div>
    );
  }

  const list = response?.data || [];
  const experiments = experimentsRes?.data || [];
  const activeExps = experiments.filter(e => e.status === 'running' || e.status === 'pending');

  // Filter list
  const filteredList = list.filter(c => {
    const matchesSearch = c.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                          c.id.toLowerCase().includes(searchTerm.toLowerCase());
    
    if (filterMode === 'monitored') return matchesSearch && c.is_monitored;
    if (filterMode === 'unmonitored') return matchesSearch && !c.is_monitored;
    return matchesSearch;
  });

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const dm = 2;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  };

  const formatUptime = (seconds: number): string => {
    if (seconds <= 0) return '0s';
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    if (hrs > 0) return `${hrs}h ${mins}m`;
    if (mins > 0) return `${mins}m ${secs}s`;
    return `${secs}s`;
  };

  const triggerAction = (container: Container, type: string) => {
    if (type === 'kill' || type === 'stop' || type === 'restart') {
      setConfirmDialog({
        open: true,
        containerId: container.id,
        containerName: container.name,
        attackType: type
      });
    } else {
      injectAttack.mutate({ id: container.id, type });
    }
  };

  return (
    <div className="space-y-6">
      
      {/* Search & Filters bar */}
      <div className="flex flex-col sm:flex-row gap-4 items-stretch sm:items-center justify-between border-b border-slate-800/80 pb-6">
        <div className="flex items-center gap-3 bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2 w-full sm:w-80">
          <Search className="h-4 w-4 text-slate-500" />
          <input 
            type="text" 
            placeholder="Search by container name/ID..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="bg-transparent text-sm border-0 focus:outline-hidden w-full text-gray-200"
          />
        </div>

        <div className="flex gap-2">
          {['all', 'monitored', 'unmonitored'].map((mode) => (
            <button
              key={mode}
              onClick={() => setFilterMode(mode as any)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold uppercase tracking-wider cursor-pointer border ${
                filterMode === mode 
                  ? 'bg-violet-600/15 text-violet-400 border-violet-500/30' 
                  : 'border-slate-800 text-slate-400 hover:bg-slate-800/50'
              }`}
            >
              {mode}
            </button>
          ))}
        </div>
      </div>

      {/* Grid of Containers */}
      {filteredList.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredList.map((container) => {
            const isMonitored = container.is_monitored;
            const activeExp = activeExps.find(e => e.target_container_id === container.id);
            
            return (
              <div 
                key={container.id} 
                className={`p-6 rounded-xl border flex flex-col justify-between transition-all duration-200 relative overflow-hidden ${
                  activeExp
                    ? 'border-rose-500 bg-rose-950/10 shadow-[0_0_15px_rgba(239,68,68,0.05)]'
                    : container.status === 'running' 
                      ? 'border-slate-800 bg-[#0f172a]/10 hover:border-slate-700/80' 
                      : 'border-rose-950/40 bg-rose-950/5'
                }`}
              >
                
                {/* Active Attack Banner Indicator */}
                {activeExp && (
                  <div className="absolute top-0 left-0 right-0 bg-rose-600/90 text-white text-[9px] font-bold uppercase tracking-widest text-center py-0.5 animate-pulse">
                    Stress Test Active: {activeExp.attack_type}
                  </div>
                )}

                {/* Upper Details */}
                <div className={activeExp ? 'pt-2' : ''}>
                  <div className="flex justify-between items-start">
                    <div className="truncate max-w-[70%]">
                      <h3 className="font-bold text-sm truncate text-gray-200" title={container.name}>{container.name}</h3>
                      <p className="text-[10px] text-slate-500 mt-0.5 truncate font-mono">{container.image}</p>
                    </div>

                    {/* Status Badge */}
                    <div className="flex flex-col items-end gap-1.5">
                      <span className={`px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider ${
                        container.status === 'running'
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : container.status === 'paused'
                            ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                            : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                      }`}>
                        {container.status}
                      </span>
                      {isMonitored ? (
                        <div className="flex items-center gap-1 text-[9px] font-semibold text-violet-400">
                          <ShieldCheck className="h-3 w-3" />
                          <span>MONITORED</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1 text-[9px] font-semibold text-slate-600">
                          <ShieldAlert className="h-3 w-3" />
                          <span>EXCLUDED</span>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Resource Stats */}
                  <div className="grid grid-cols-3 gap-2 mt-6 text-xs text-slate-400">
                    <div className="flex flex-col gap-1">
                      <span className="text-[10px] text-slate-500 flex items-center gap-1 uppercase font-bold tracking-wider">
                        <Cpu className="h-3.5 w-3.5 text-slate-600" /> CPU
                      </span>
                      <span className="font-bold text-gray-200">{container.cpu_usage.toFixed(1)}%</span>
                    </div>
                    <div className="flex flex-col gap-1">
                      <span className="text-[10px] text-slate-500 flex items-center gap-1 uppercase font-bold tracking-wider">
                        <Database className="h-3.5 w-3.5 text-slate-600" /> RAM
                      </span>
                      <span className="font-bold text-gray-200">{formatBytes(container.memory_usage)}</span>
                    </div>
                    <div className="flex flex-col gap-1">
                      <span className="text-[10px] text-slate-500 flex items-center gap-1 uppercase font-bold tracking-wider">
                        <Clock className="h-3.5 w-3.5 text-slate-600" /> UPTIME
                      </span>
                      <span className="font-bold text-gray-200">{formatUptime(container.uptime)}</span>
                    </div>
                  </div>
                </div>

                {/* Container control inputs */}
                {isMonitored && (
                  <div className="mt-6 pt-4 border-t border-slate-800/80 grid grid-cols-4 gap-2">
                    
                    {/* Pause/Unpause */}
                    {container.status === 'paused' ? (
                      <button 
                        onClick={() => triggerAction(container, 'unpause')}
                        className="py-2 px-1 flex flex-col items-center justify-center gap-1 rounded-lg text-[9px] font-bold uppercase tracking-wider text-emerald-400 bg-emerald-500/5 hover:bg-emerald-500/15 cursor-pointer transition-colors border border-transparent hover:border-emerald-500/20"
                        title="Resume Container Execution"
                      >
                        <Play className="h-3.5 w-3.5" />
                        <span>Resume</span>
                      </button>
                    ) : (
                      <button 
                        onClick={() => triggerAction(container, 'pause')}
                        disabled={container.status !== 'running'}
                        className="py-2 px-1 flex flex-col items-center justify-center gap-1 rounded-lg text-[9px] font-bold uppercase tracking-wider text-amber-400 bg-amber-500/5 hover:bg-amber-500/15 cursor-pointer transition-colors border border-transparent hover:border-amber-500/20 disabled:opacity-30 disabled:pointer-events-none"
                        title="Freeze Container Threads"
                      >
                        <Pause className="h-3.5 w-3.5" />
                        <span>Pause</span>
                      </button>
                    )}

                    {/* Stop */}
                    <button 
                      onClick={() => triggerAction(container, 'stop')}
                      disabled={container.status !== 'running'}
                      className="py-2 px-1 flex flex-col items-center justify-center gap-1 rounded-lg text-[9px] font-bold uppercase tracking-wider text-rose-400 bg-rose-500/5 hover:bg-rose-500/15 cursor-pointer transition-colors border border-transparent hover:border-rose-500/20 disabled:opacity-30 disabled:pointer-events-none"
                      title="Shut Down Container Process"
                    >
                      <Square className="h-3.5 w-3.5" />
                      <span>Stop</span>
                    </button>

                    {/* Restart */}
                    <button 
                      onClick={() => triggerAction(container, 'restart')}
                      disabled={container.status !== 'running'}
                      className="py-2 px-1 flex flex-col items-center justify-center gap-1 rounded-lg text-[9px] font-bold uppercase tracking-wider text-violet-400 bg-violet-500/5 hover:bg-violet-500/15 cursor-pointer transition-colors border border-transparent hover:border-violet-500/20 disabled:opacity-30 disabled:pointer-events-none"
                      title="Perform Restart Loop"
                    >
                      <RefreshCw className="h-3.5 w-3.5" />
                      <span>Restart</span>
                    </button>

                    {/* Kill */}
                    <button 
                      onClick={() => triggerAction(container, 'kill')}
                      disabled={container.status !== 'running'}
                      className="py-2 px-1 flex flex-col items-center justify-center gap-1 rounded-lg text-[9px] font-bold uppercase tracking-wider text-red-500 bg-red-500/5 hover:bg-red-500/15 cursor-pointer transition-colors border border-transparent hover:border-red-500/20 disabled:opacity-30 disabled:pointer-events-none"
                      title="Kill Container Process instantly"
                    >
                      <Skull className="h-3.5 w-3.5" />
                      <span>Kill</span>
                    </button>

                  </div>
                )}

              </div>
            );
          })}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center min-h-[40vh] border border-slate-800/80 rounded-xl bg-slate-900/10 p-6 text-center">
          <Search className="h-8 w-8 text-slate-600 mb-2" />
          <h3 className="font-semibold text-sm">No containers found</h3>
          <p className="text-xs text-slate-500 mt-1">Try clearing filters or check another search term.</p>
        </div>
      )}

      {/* Dangerous Operations Modal Dialog Confirmation Overlay */}
      {confirmDialog && confirmDialog.open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4">
          <div className="w-full max-w-md p-6 rounded-xl border border-slate-850 bg-[#0f172a] shadow-2xl flex flex-col gap-4">
            <div className="flex items-center gap-3 text-amber-500">
              <AlertTriangle className="h-6 w-6" />
              <h3 className="font-bold text-base">Confirm Dangerous Action</h3>
            </div>
            
            <p className="text-sm text-slate-350 leading-relaxed">
              Are you sure you want to trigger a **{confirmDialog.attackType.toUpperCase()}** failure injection against container **{confirmDialog.containerName}**? 
              This will disrupt running services.
            </p>

            <div className="flex items-center justify-end gap-3 mt-4">
              <button
                onClick={() => setConfirmDialog(null)}
                className="px-4 py-2 border border-slate-800 text-slate-400 rounded-lg text-xs font-bold uppercase tracking-wider hover:bg-slate-800/50 cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={() => injectAttack.mutate({ id: confirmDialog.containerId, type: confirmDialog.attackType })}
                disabled={injectAttack.isPending}
                className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs font-bold uppercase tracking-wider cursor-pointer flex items-center gap-1.5 disabled:opacity-50"
              >
                {injectAttack.isPending ? 'Executing...' : 'Confirm Disruption'}
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}

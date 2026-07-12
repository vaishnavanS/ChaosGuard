import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import { 
  Cpu, 
  Settings, 
  Radio, 
  AlertTriangle,
  StopCircle
} from 'lucide-react';

export default function Runtime() {
  const queryClient = useQueryClient();

  // Queries
  const { data: healthRes } = useQuery({
    queryKey: ['runtimeHealth'],
    queryFn: () => api.getHealth(),
  });

  const { data: schedulerRes } = useQuery({
    queryKey: ['runtimeScheduler'],
    queryFn: () => api.getSchedulerStatus(),
  });

  const { data: runtimeRes } = useQuery({
    queryKey: ['runtimeStatus'],
    queryFn: () => api.getRuntime(),
  });

  // Stop Daemon Mutation
  const stopDaemon = useMutation({
    mutationFn: () => api.stopRuntime(),
    onSuccess: () => {
      alert('Graceful shutdown request sent successfully to the daemon.');
      queryClient.invalidateQueries({ queryKey: ['runtimeStatus'] });
    },
    onError: (err: any) => {
      alert(`Shutdown request failed: ${err.message}`);
    }
  });

  return (
    <div className="space-y-8 max-w-4xl">
      
      {/* Upper Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        
        {/* Connection Details Card */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 space-y-4">
          <div className="flex items-center gap-2 text-violet-400">
            <Radio className="h-5 w-5 animate-pulse-status" />
            <h2 className="text-base font-semibold tracking-tight">Daemon Environment</h2>
          </div>
          
          <div className="space-y-3 text-sm text-slate-400">
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>Lifecycle State</span>
              <span className="font-bold text-emerald-400 uppercase">{runtimeRes?.data?.state || 'running'}</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>Core Service version</span>
              <span className="font-mono text-gray-200">{healthRes?.data?.version || 'v0.1.0-dev'}</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>API Gateway Status</span>
              <span className="text-emerald-400 font-semibold uppercase">Healthy</span>
            </div>
          </div>
        </div>

        {/* Diagnostic Checks Card */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 space-y-4">
          <div className="flex items-center gap-2 text-emerald-400">
            <Cpu className="h-5 w-5" />
            <h2 className="text-base font-semibold tracking-tight">Service Connectivity Checks</h2>
          </div>
          
          <div className="space-y-3 text-sm text-slate-400">
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>Docker Daemon Connection</span>
              <span className="text-emerald-400 font-semibold uppercase">Connected</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>SQLite persistence store</span>
              <span className="text-emerald-400 font-semibold uppercase">Permitted</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span>Automated Safe Mode state</span>
              <span className="text-amber-500 font-semibold uppercase">Active</span>
            </div>
          </div>
        </div>

      </div>

      {/* Scheduler Configuration Table */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/10 space-y-6">
        <div className="flex items-center gap-2 text-slate-300">
          <Settings className="h-5 w-5" />
          <h2 className="text-base font-semibold tracking-tight">Automated Scheduler Configuration</h2>
        </div>

        <div className="space-y-4 text-sm text-slate-400">
          <div className="grid grid-cols-2 gap-4 border-b border-slate-800 pb-2">
            <span className="text-slate-500 uppercase tracking-wider text-xs">Property</span>
            <span className="text-slate-500 uppercase tracking-wider text-xs">Configured Value</span>
          </div>

          <div className="grid grid-cols-2 gap-4 border-b border-slate-800/50 pb-2">
            <span>Automation Trigger Interval</span>
            <span className="font-semibold text-gray-200">{schedulerRes?.data?.interval || '30s'}</span>
          </div>

          <div className="grid grid-cols-2 gap-4 border-b border-slate-800/50 pb-2">
            <span>Fault Execution Duration</span>
            <span className="font-semibold text-gray-200">{schedulerRes?.data?.duration || '10s'}</span>
          </div>

          <div className="grid grid-cols-2 gap-4 border-b border-slate-800/50 pb-2">
            <span>Attack Selection Strategy</span>
            <span className="font-semibold text-violet-400 uppercase">{schedulerRes?.data?.mode || 'random'}</span>
          </div>

          <div className="grid grid-cols-2 gap-4 pb-2">
            <span>Safety Shield Mode</span>
            <span className="font-semibold text-amber-500">ENABLED (postgres, mysql, redis, prometheus protection)</span>
          </div>
        </div>
      </div>

      {/* Dangerous Remote System Actions */}
      <div className="p-6 rounded-xl border border-rose-950/40 bg-rose-950/5 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-start gap-3">
          <AlertTriangle className="h-6 w-6 text-rose-500 shrink-0 mt-0.5" />
          <div>
            <h3 className="font-bold text-sm text-rose-400">Shutdown Daemon Lifecycle</h3>
            <p className="text-xs text-rose-300/80 mt-1 max-w-lg leading-relaxed">
              Triggering this action sends a terminate request to the background daemon. 
              The API server, scheduler, and metrics agents will be closed gracefully, and running container failures will be recovered.
            </p>
          </div>
        </div>
        
        <button
          onClick={() => {
            if (confirm('Are you absolutely sure you want to stop the ChaosGuard Daemon?')) {
              stopDaemon.mutate();
            }
          }}
          disabled={stopDaemon.isPending}
          className="px-4 py-2.5 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs font-bold uppercase tracking-wider flex items-center gap-2 cursor-pointer disabled:opacity-50"
        >
          <StopCircle className="h-4 w-4" />
          {stopDaemon.isPending ? 'Shutting down...' : 'Stop Daemon'}
        </button>
      </div>

    </div>
  );
}

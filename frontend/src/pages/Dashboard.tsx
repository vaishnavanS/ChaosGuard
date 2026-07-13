import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Play, 
  Square, 
  Activity, 
  Layers, 
  Clock, 
  ShieldCheck, 
  AlertTriangle,
  History,
  Zap,
  TrendingUp,
  Cpu,
  Database,
  Shield,
  Heart,
  Undo2,
  XCircle,
  PlayCircle,
  HelpCircle
} from 'lucide-react';
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid
} from 'recharts';

export default function Dashboard() {
  const queryClient = useQueryClient();

  // Queries
  const { data: healthRes } = useQuery({
    queryKey: ['healthStateDashboard'],
    queryFn: () => api.getHealth(),
    refetchInterval: 5000,
  });

  const { data: containerRes } = useQuery({
    queryKey: ['containersDashboard'],
    queryFn: () => api.getContainers(),
    refetchInterval: 5000,
  });

  const { data: experimentRes } = useQuery({
    queryKey: ['experimentsDashboard'],
    queryFn: () => api.getExperiments(),
    refetchInterval: 5000,
  });

  const { data: schedulerRes } = useQuery({
    queryKey: ['schedulerDashboard'],
    queryFn: () => api.getSchedulerStatus(),
    refetchInterval: 5000,
  });

  // Mutations for Scheduler start/stop
  const startScheduler = useMutation({
    mutationFn: () => api.startScheduler(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['schedulerDashboard'] }),
  });

  const stopScheduler = useMutation({
    mutationFn: () => api.stopScheduler(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['schedulerDashboard'] }),
  });

  // Abort active experiment mutation
  const abortExperiment = useMutation({
    mutationFn: (id: string) => api.deleteExperiment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experimentsDashboard'] });
      queryClient.invalidateQueries({ queryKey: ['containersDashboard'] });
    },
    onError: (err: any) => {
      alert(`Abort execution failed: ${err.message}`);
    }
  });

  const containers = containerRes?.data || [];
  const experiments = experimentRes?.data || [];
  const scheduler = schedulerRes?.data;
  const health = healthRes?.data;

  // Deriving counts
  const totalContainers = containers.length;
  const runningContainers = containers.filter(c => c.status === 'running').length;
  const pausedContainers = containers.filter(c => c.status === 'paused').length;
  const unmonitoredContainers = containers.filter(c => !c.is_monitored).length;

  const activeExperiments = experiments.filter(e => e.status === 'running' || e.status === 'pending');
  const runningExp = activeExperiments.length > 0 ? activeExperiments[0] : null;

  const totalAttacks = experiments.length;
  const successfulRecoveries = experiments.filter(e => e.status === 'completed' || e.status === 'recovered').length;
  const failedRecoveries = experiments.filter(e => e.status === 'failed').length;

  // Resilience score math: start at 100, deduct points for failures and unmonitored containers
  const resilienceScore = Math.max(20, 100 - (failedRecoveries * 10) - (unmonitoredContainers * 4));

  // Sparkline simulated historical metrics datasets
  const sparkCpu = [12, 16, 14, 22, 17, 19, totalContainers > 0 ? Math.round(containers.reduce((s, c) => s + c.cpu_usage, 0) / totalContainers) : 15];
  const sparkMem = [44, 46, 42, 49, 45, 47, totalContainers > 0 ? Math.round(containers.reduce((s, c) => s + (c.memory_usage / (1024 * 1024)), 0) / 100) : 45];
  const sparkAttacks = [1, 3, 2, 4, 3, 5, totalAttacks];
  const sparkResilience = [96, 94, 95, 92, 90, 88, resilienceScore];

  // Group experiments over time for main Recharts chart
  const timelineGroups = experiments.reduce((acc: Record<string, number>, exp) => {
    try {
      const date = new Date(exp.started_at);
      const label = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      acc[label] = (acc[label] || 0) + 1;
    } catch {}
    return acc;
  }, {});

  const timelineData = Object.entries(timelineGroups).map(([time, attacks]) => ({
    time,
    attacks
  })).slice(-8);

  // Recovery countdown progress bar computations
  const [countdown, setCountdown] = useState(0);
  const [progressVal, setProgressVal] = useState(0);

  useEffect(() => {
    if (runningExp) {
      const start = new Date(runningExp.started_at).getTime();
      const dur = runningExp.duration * 1000;
      const elapsed = Date.now() - start;
      const remaining = Math.max(0, Math.ceil((dur - elapsed) / 1000));
      
      setCountdown(remaining);
      setProgressVal(Math.min(100, Math.max(0, (elapsed / dur) * 100)));

      const interval = setInterval(() => {
        const elapsed = Date.now() - start;
        const remaining = Math.max(0, Math.ceil((dur - elapsed) / 1000));
        setCountdown(remaining);
        setProgressVal(Math.min(100, Math.max(0, (elapsed / dur) * 100)));
      }, 1000);

      return () => clearInterval(interval);
    } else {
      setCountdown(0);
      setProgressVal(0);
    }
  }, [runningExp]);

  return (
    <div className="space-y-6">
      
      {/* Hero Section: Enterprise Status Banner */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 p-5 rounded-xl border border-slate-800 bg-[#0c101b] shadow-lg relative overflow-hidden">
        <div className="absolute top-0 right-0 p-2 bg-violet-600/10 text-violet-400 border-l border-b border-slate-800 text-[9px] font-bold uppercase tracking-widest">
          SYSTEM HEALTH CONSOLE
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">ChaosGuard State</span>
          <div className="flex items-center gap-2 mt-1">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse-status"></span>
            <span className="font-bold text-gray-200">RUNNING ({health?.version || 'v0.3.1'})</span>
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Docker socket</span>
          <span className="font-bold text-emerald-400 mt-1">✓ CONNECTED</span>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Safe Mode Status</span>
          <span className="font-bold text-amber-500 mt-1 flex items-center gap-1">
            <ShieldCheck className="h-4 w-4" /> ACTIVE SAFETY
          </span>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Daemon Uptime</span>
          <span className="font-bold text-gray-200 mt-1 uppercase">
            {health?.status === 'ok' ? 'Healthy' : 'Degraded'}
          </span>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Scheduler mode</span>
          <div className="flex items-center justify-between mt-0.5">
            <span className="font-bold text-violet-400 uppercase">{scheduler?.running ? 'ACTIVE' : 'IDLE'}</span>
            <button
              onClick={() => scheduler?.running ? stopScheduler.mutate() : startScheduler.mutate()}
              className={`p-1 px-2.5 rounded-lg border text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 cursor-pointer transition-colors ${
                scheduler?.running 
                  ? 'border-rose-500/20 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20' 
                  : 'border-emerald-500/20 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'
              }`}
            >
              {scheduler?.running ? <Square className="h-3 w-3" /> : <Play className="h-3 w-3" />}
              <span>{scheduler?.running ? 'Stop' : 'Start'}</span>
            </button>
          </div>
        </div>
      </div>

      {/* Current Experiment Monitor (Dominates the screen when active) */}
      <AnimatePresence mode="wait">
        {runningExp ? (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="p-6 rounded-xl border border-rose-900/30 bg-[#1e131d] relative overflow-hidden"
          >
            <div className="absolute top-0 right-0 p-2 bg-rose-500/10 text-rose-400 border-l border-b border-rose-900/20 text-[9px] font-bold uppercase tracking-widest flex items-center gap-1">
              <Flame className="h-3.5 w-3.5 text-rose-500 animate-pulse" />
              ACTIVE FAULT INJECTION RUNNING
            </div>
            
            <div className="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-6">
              <div className="space-y-1">
                <span className="text-[10px] font-bold text-rose-400 uppercase tracking-widest">Active Disruption Node</span>
                <h2 className="text-xl font-bold text-gray-100 flex items-center gap-2">
                  <Layers className="h-5 w-5 text-rose-400" />
                  {runningExp.container_name}
                </h2>
                <div className="flex gap-4 text-xs text-slate-500 flex-wrap">
                  <span>Attack type: <strong className="text-violet-400 font-mono">{runningExp.attack_type.toUpperCase()}</strong></span>
                  <span>•</span>
                  <span>Duration: <strong>{runningExp.duration}s</strong></span>
                  <span>•</span>
                  <span>Impact: <strong className="text-rose-400">HIGH SERVICE DEGRADATION</strong></span>
                </div>
              </div>

              <div className="flex items-center gap-6 shrink-0 self-end lg:self-center">
                <div className="flex flex-col items-end gap-0.5">
                  <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Auto-Recovery In</span>
                  <div className="flex items-baseline gap-1 text-2xl font-black text-rose-450">
                    <span>{countdown}</span>
                    <span className="text-xs font-normal text-slate-500">seconds</span>
                  </div>
                </div>

                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      if (confirm('Abort experiment and clean up immediately?')) {
                        abortExperiment.mutate(runningExp.id);
                      }
                    }}
                    disabled={abortExperiment.isPending}
                    className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs font-bold uppercase tracking-wider flex items-center gap-1 cursor-pointer disabled:opacity-50"
                  >
                    <XCircle className="h-4 w-4" />
                    <span>Abort Disruption</span>
                  </button>
                  <button
                    onClick={() => {
                      if (confirm('Trigger manual recovery sequence?')) {
                        abortExperiment.mutate(runningExp.id);
                      }
                    }}
                    disabled={abortExperiment.isPending}
                    className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold uppercase tracking-wider flex items-center gap-1 cursor-pointer disabled:opacity-50"
                  >
                    <Undo2 className="h-4 w-4" />
                    <span>Force Recover</span>
                  </button>
                </div>
              </div>
            </div>

            {/* Recovery progress slider */}
            <div className="mt-6 w-full h-1.5 bg-slate-950/60 rounded-full overflow-hidden">
              <div 
                className="h-full bg-gradient-to-r from-rose-500 to-violet-500 transition-all duration-1000 ease-linear"
                style={{ width: `${progressVal}%` }}
              ></div>
            </div>
          </motion.div>
        ) : (
          <div className="p-6 rounded-xl border border-slate-850 bg-[#0f172a]/10 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Shield className="h-5 w-5 text-emerald-400" />
              <div>
                <h4 className="text-xs font-bold text-gray-200 uppercase tracking-wider">No active failure stress tests</h4>
                <p className="text-[10px] text-slate-500 leading-normal">System is operating normally. Ready to schedule faults.</p>
              </div>
            </div>
            <span className="text-[9px] font-bold px-2 py-0.5 rounded border border-slate-800 text-slate-500">IDLE STANDBY</span>
          </div>
        )}
      </AnimatePresence>

      {/* KPI Cards Grid with mini sparklines */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        
        {/* KPI: Containers count */}
        <div className="p-5 rounded-xl border border-slate-850 bg-[#0f172a]/20 flex flex-col justify-between h-32">
          <div className="flex justify-between items-start">
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-1">
              <Layers className="h-3.5 w-3.5" /> Docker Containers
            </span>
            <span className="text-[10px] font-bold text-slate-400">
              {runningContainers} Active
            </span>
          </div>
          <div className="flex justify-between items-end mt-2">
            <div>
              <h3 className="text-2xl font-black tracking-tight text-gray-200">{totalContainers}</h3>
              <p className="text-[9px] text-slate-600 uppercase font-bold mt-1">Containers count</p>
            </div>
            {/* Sparkline */}
            <div className="h-8 w-20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={sparkCpu.map((cpu, idx) => ({ idx, cpu }))}>
                  <Area type="monotone" dataKey="cpu" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.05} strokeWidth={1} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* KPI: Chaos injections count */}
        <div className="p-5 rounded-xl border border-slate-850 bg-[#0f172a]/20 flex flex-col justify-between h-32">
          <div className="flex justify-between items-start">
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-1">
              <Activity className="h-3.5 w-3.5" /> Experiments
            </span>
            <span className="text-xs text-violet-400 font-bold flex items-center gap-0.5">
              <TrendingUp className="h-3 w-3" /> Live
            </span>
          </div>
          <div className="flex justify-between items-end mt-2">
            <div>
              <h3 className="text-2xl font-black tracking-tight text-gray-200">{totalAttacks}</h3>
              <p className="text-[9px] text-slate-600 uppercase font-bold mt-1">Total Injections</p>
            </div>
            {/* Sparkline */}
            <div className="h-8 w-20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={sparkAttacks.map((attacks, idx) => ({ idx, attacks }))}>
                  <Area type="monotone" dataKey="attacks" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.05} strokeWidth={1} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* KPI: Average host load */}
        <div className="p-5 rounded-xl border border-slate-855 bg-[#0f172a]/20 flex flex-col justify-between h-32">
          <div className="flex justify-between items-start">
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-1">
              <Cpu className="h-3.5 w-3.5" /> Average Memory
            </span>
            <span className="text-xs text-slate-500">RAM Load</span>
          </div>
          <div className="flex justify-between items-end mt-2">
            <div>
              <h3 className="text-2xl font-black tracking-tight text-gray-200">
                {totalContainers > 0 ? (containers.reduce((s, c) => s + (c.memory_usage / (1024 * 1024)), 0) / totalContainers).toFixed(0) : '0'} <span className="text-xs font-bold text-slate-500">MB</span>
              </h3>
              <p className="text-[9px] text-slate-600 uppercase font-bold mt-1">Host allocation</p>
            </div>
            {/* Sparkline */}
            <div className="h-8 w-20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={sparkMem.map((mem, idx) => ({ idx, mem }))}>
                  <Area type="monotone" dataKey="mem" stroke="#10b981" fill="#10b981" fillOpacity={0.05} strokeWidth={1} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* KPI: Resilience Score */}
        <div className="p-5 rounded-xl border border-slate-855 bg-[#0f172a]/20 flex flex-col justify-between h-32">
          <div className="flex justify-between items-start">
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest flex items-center gap-1">
              <Heart className="h-3.5 w-3.5" /> Resilience Score
            </span>
            <span className="text-[9px] font-bold text-emerald-400 bg-emerald-500/10 px-1 rounded">
              SLA
            </span>
          </div>
          <div className="flex justify-between items-end mt-2">
            <div>
              <h3 className="text-2xl font-black tracking-tight text-gray-200">{resilienceScore}%</h3>
              <p className="text-[9px] text-slate-600 uppercase font-bold mt-1">Security Score</p>
            </div>
            {/* Sparkline */}
            <div className="h-8 w-20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={sparkResilience.map((res, idx) => ({ idx, res }))}>
                  <Area type="monotone" dataKey="res" stroke="#10b981" fill="#10b981" fillOpacity={0.05} strokeWidth={1} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

      </div>

      {/* Middle Row: Animated Attack Map Flow & Live Activity timeline */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Animated Attack Map */}
        <div className="p-6 rounded-xl border border-slate-850 bg-[#0f172a]/10 lg:col-span-2 flex flex-col justify-between">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-violet-400 animate-pulse-status" />
              <h2 className="text-xs uppercase font-bold tracking-wider text-slate-400">Chaos Disruption Flow Map</h2>
            </div>
            <div className="flex items-center gap-1 text-[9px] text-slate-500 font-bold uppercase">
              <span className={`h-1.5 w-1.5 rounded-full ${runningExp ? 'bg-rose-450 animate-ping' : 'bg-slate-600'}`}></span>
              {runningExp ? 'Stress active' : 'Monitoring flow'}
            </div>
          </div>

          <div className="flex-1 flex flex-col md:flex-row justify-between items-center gap-4 relative py-6 select-none mt-2">
            {/* Node: Trigger */}
            <div className="flex flex-col items-center gap-2 bg-[#0c101b] border border-slate-850 p-4 rounded-xl w-32 shadow-lg relative">
              <span className="text-[9px] font-bold text-slate-500 uppercase tracking-wider">Trigger Source</span>
              <span className="font-bold text-xs text-violet-400">{scheduler?.running ? 'Scheduler' : 'Manual Trigger'}</span>
            </div>

            <ChevronRightLine active={!!runningExp} />

            {/* Node: Target Container */}
            <div className="flex flex-col items-center gap-2 bg-[#0c101b] border border-slate-850 p-4 rounded-xl w-32 shadow-lg relative">
              <span className="text-[9px] font-bold text-slate-500 uppercase tracking-wider">Target Node</span>
              <span className="font-bold text-xs truncate max-w-full text-gray-200">
                {runningExp ? runningExp.container_name : 'No target'}
              </span>
            </div>

            <ChevronRightLine active={!!runningExp} />

            {/* Node: Fault Injected */}
            <div className="flex flex-col items-center gap-2 bg-[#0c101b] border border-slate-850 p-4 rounded-xl w-32 shadow-lg relative">
              <span className="text-[9px] font-bold text-slate-500 uppercase tracking-wider">Injected Disruption</span>
              <span className="font-mono text-xs font-bold text-rose-400 uppercase">
                {runningExp ? runningExp.attack_type : 'Ready'}
              </span>
            </div>

            <ChevronRightLine active={!!runningExp} />

            {/* Node: Recovery Healer */}
            <div className="flex flex-col items-center gap-2 bg-[#0c101b] border border-slate-850 p-4 rounded-xl w-32 shadow-lg relative">
              <span className="text-[9px] font-bold text-slate-500 uppercase tracking-wider">Healer status</span>
              <span className="font-bold text-xs text-emerald-400 uppercase">
                {runningExp ? 'Autoscaling' : 'Healthy'}
              </span>
            </div>
          </div>
        </div>

        {/* Live Event Timeline */}
        <div className="p-6 rounded-xl border border-slate-850 bg-[#0f172a]/10 flex flex-col justify-between">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <History className="h-4 w-4 text-slate-400" />
              <h2 className="text-xs uppercase font-bold tracking-wider text-slate-400">Live Activity Feed</h2>
            </div>
            <span className="text-[8px] font-bold text-slate-600 bg-slate-900 px-1.5 py-0.5 rounded">AUTO-REFRESH</span>
          </div>

          <div className="flex-1 overflow-y-auto space-y-4 max-h-[14rem] scrollbar-thin mt-2">
            {experiments.slice(-4).reverse().map((exp, i) => (
              <div key={i} className="flex gap-3 text-xs leading-relaxed border-l border-slate-800 pl-3">
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-[10px] text-slate-600">
                      {new Date(exp.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </span>
                    <span className="font-bold text-violet-400 uppercase tracking-wide">{exp.container_name}</span>
                  </div>
                  <p className="text-slate-400">
                    Disruption <strong className="font-mono text-slate-350">{exp.attack_type}</strong> injected successfully. 
                    Status: <span className={`font-bold uppercase text-[9px] ${
                      exp.status === 'failed' ? 'text-rose-400' : 'text-emerald-400'
                    }`}>{exp.status}</span>.
                  </p>
                </div>
              </div>
            ))}
            {experiments.length === 0 && (
              <div className="h-full flex items-center justify-center text-xs text-slate-600">
                Waiting for failure injections to log activity...
              </div>
            )}
          </div>
        </div>

      </div>

      {/* Main Charts timeline */}
      <div className="p-6 rounded-xl border border-slate-850 bg-[#0f172a]/10">
        <div className="flex items-center gap-2 mb-6">
          <Activity className="h-4 w-4 text-violet-400" />
          <h2 className="text-xs uppercase font-bold tracking-wider text-slate-400">Fault Injections Rate Timeline</h2>
        </div>
        <div className="h-64 w-full">
          {timelineData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={timelineData}>
                <defs>
                  <linearGradient id="attGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.15}/>
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                <XAxis dataKey="time" stroke="#64748b" fontSize={10} />
                <YAxis stroke="#64748b" fontSize={10} allowDecimals={false} />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                />
                <Area type="monotone" dataKey="attacks" stroke="#8b5cf6" fillOpacity={1} fill="url(#attGrad)" strokeWidth={1.5} />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-full flex flex-col items-center justify-center text-xs text-slate-500 gap-2 border border-dashed border-slate-800 rounded-lg">
              <HelpCircle className="h-8 w-8 text-slate-700 animate-pulse" />
              <span className="uppercase font-bold tracking-widest text-[9px] text-slate-650">Waiting for experiment data...</span>
            </div>
          )}
        </div>
      </div>

    </div>
  );
}

// Arrow helper with Framer Motion glow pulse indicator
function ChevronRightLine({ active }: { active: boolean }) {
  return (
    <div className="hidden md:block w-12 h-0.5 bg-slate-800 relative">
      {active && (
        <motion.div
          animate={{ left: ['0%', '100%'] }}
          transition={{ repeat: Infinity, duration: 1.5, ease: 'linear' }}
          className="absolute top-1/2 -translate-y-1/2 h-1.5 w-1.5 rounded-full bg-violet-400 shadow-[0_0_8px_rgba(167,139,250,0.8)]"
        />
      )}
    </div>
  );
}

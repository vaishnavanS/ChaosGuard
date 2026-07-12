import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import { 
  Play, 
  Square, 
  Activity, 
  Layers, 
  CheckCircle2, 
  XCircle, 
  AlertTriangle,
  RotateCcw,
  Clock
} from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend
} from 'recharts';

export default function Dashboard() {
  const queryClient = useQueryClient();

  // Queries
  const { data: healthRes, isLoading: hLoading } = useQuery({
    queryKey: ['health'],
    queryFn: () => api.getHealth(),
    refetchInterval: 5000,
  });

  const { data: containerRes, isLoading: cLoading } = useQuery({
    queryKey: ['containers'],
    queryFn: () => api.getContainers(),
    refetchInterval: 5000,
  });

  const { data: experimentRes, isLoading: eLoading } = useQuery({
    queryKey: ['experiments'],
    queryFn: () => api.getExperiments(),
    refetchInterval: 5000,
  });

  const { data: schedulerRes, isLoading: sLoading } = useQuery({
    queryKey: ['scheduler'],
    queryFn: () => api.getSchedulerStatus(),
    refetchInterval: 5000,
  });

  // Mutations
  const startScheduler = useMutation({
    mutationFn: () => api.startScheduler(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scheduler'] });
    }
  });

  const stopScheduler = useMutation({
    mutationFn: () => api.stopScheduler(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scheduler'] });
    }
  });

  if (hLoading || cLoading || eLoading || sLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-violet-500"></div>
        <p className="text-sm text-gray-400">Loading system metrics...</p>
      </div>
    );
  }

  // Derive counts
  const containers = containerRes?.data || [];
  const experiments = experimentRes?.data || [];
  const scheduler = schedulerRes?.data;

  const totalContainers = containers.length;
  const runningContainers = containers.filter(c => c.status === 'running').length;
  const pausedContainers = containers.filter(c => c.status === 'paused').length;
  const inactiveContainers = totalContainers - runningContainers - pausedContainers;

  const totalExperiments = experiments.length;
  const activeExperiments = experiments.filter(e => e.status === 'running' || e.status === 'pending').length;
  const completedExperiments = experiments.filter(e => e.status === 'completed' || e.status === 'recovered').length;
  const failedExperiments = experiments.filter(e => e.status === 'failed').length;

  const attackCounts = experiments.reduce((acc: Record<string, number>, exp) => {
    acc[exp.attack_type] = (acc[exp.attack_type] || 0) + 1;
    return acc;
  }, {});

  const attackData = Object.entries(attackCounts).map(([name, value]) => ({
    name: name.toUpperCase(),
    value
  }));

  // Group experiments over time (e.g. by hour/date)
  const timeGroups = experiments.reduce((acc: Record<string, number>, exp) => {
    try {
      const date = new Date(exp.started_at);
      const label = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      acc[label] = (acc[label] || 0) + 1;
    } catch {
      // Ignored
    }
    return acc;
  }, {});

  const chartData = Object.entries(timeGroups).map(([time, count]) => ({
    time,
    attacks: count
  })).slice(-8); // Get last 8 buckets

  const containerData = [
    { name: 'Running', value: runningContainers, color: '#10b981' },
    { name: 'Paused', value: pausedContainers, color: '#f59e0b' },
    { name: 'Stopped/Exited', value: inactiveContainers, color: '#ef4444' }
  ].filter(item => item.value > 0);

  const COLORS = ['#8b5cf6', '#3b82f6', '#10b981', '#f59e0b', '#ef4444'];

  return (
    <div className="space-y-8">
      {/* Overview Stats Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        
        {/* Runtime State Card */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/20 backdrop-blur-sm flex flex-col justify-between">
          <div className="flex justify-between items-start">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">System State</span>
            <Activity className="h-5 w-5 text-violet-400" />
          </div>
          <div className="mt-4">
            <h3 className="text-2xl font-bold tracking-tight">{healthRes?.data?.state?.toUpperCase() || 'UNKNOWN'}</h3>
            <p className="text-xs text-slate-500 mt-1">Version: {healthRes?.data?.version || 'dev'}</p>
          </div>
        </div>

        {/* Scheduler Status Card */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/20 backdrop-blur-sm flex flex-col justify-between">
          <div className="flex justify-between items-start">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Scheduler</span>
            <Clock className="h-5 w-5 text-amber-400" />
          </div>
          <div className="mt-4 flex items-center justify-between">
            <div>
              <h3 className="text-2xl font-bold tracking-tight">{scheduler?.running ? 'RUNNING' : 'STOPPED'}</h3>
              <p className="text-xs text-slate-500 mt-1">Mode: {scheduler?.mode || 'none'}</p>
            </div>
            {scheduler?.running ? (
              <button 
                onClick={() => stopScheduler.mutate()}
                disabled={stopScheduler.isPending}
                className="p-2 rounded-lg bg-rose-500/10 text-rose-400 border border-rose-500/25 hover:bg-rose-500/20 cursor-pointer disabled:opacity-50"
                title="Pause Automation"
              >
                <Square className="h-4 w-4 fill-current" />
              </button>
            ) : (
              <button 
                onClick={() => startScheduler.mutate()}
                disabled={startScheduler.isPending}
                className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/25 hover:bg-emerald-500/20 cursor-pointer disabled:opacity-50"
                title="Resume Automation"
              >
                <Play className="h-4 w-4 fill-current" />
              </button>
            )}
          </div>
        </div>

        {/* Monitored Containers Card */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/20 backdrop-blur-sm flex flex-col justify-between">
          <div className="flex justify-between items-start">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Containers</span>
            <Layers className="h-5 w-5 text-emerald-400" />
          </div>
          <div className="mt-4">
            <h3 className="text-2xl font-bold tracking-tight">{runningContainers} <span className="text-sm font-normal text-slate-500">/ {totalContainers} Running</span></h3>
            <p className="text-xs text-slate-500 mt-1">Total discovered on Docker host</p>
          </div>
        </div>

        {/* Chaos Experiments Activity Card */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/20 backdrop-blur-sm flex flex-col justify-between">
          <div className="flex justify-between items-start">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Chaos Attacks</span>
            <AlertTriangle className="h-5 w-5 text-rose-400" />
          </div>
          <div className="mt-4">
            <h3 className="text-2xl font-bold tracking-tight">{totalExperiments} <span className="text-sm font-normal text-slate-500">({activeExperiments} Active)</span></h3>
            <p className="text-xs text-slate-500 mt-1">Total faults injected since startup</p>
          </div>
        </div>

      </div>

      {/* Detail Stats Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-5 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <p className="text-xs text-slate-500 uppercase font-semibold">Successful Recoveries</p>
            <h4 className="text-xl font-bold mt-1 text-emerald-400">{completedExperiments}</h4>
          </div>
          <CheckCircle2 className="h-8 w-8 text-emerald-500/20" />
        </div>
        <div className="p-5 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <p className="text-xs text-slate-500 uppercase font-semibold">Failed Recoveries</p>
            <h4 className="text-xl font-bold mt-1 text-rose-400">{failedExperiments}</h4>
          </div>
          <XCircle className="h-8 w-8 text-rose-500/20" />
        </div>
        <div className="p-5 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <p className="text-xs text-slate-500 uppercase font-semibold">Success Rate</p>
            <h4 className="text-xl font-bold mt-1 text-violet-400">
              {totalExperiments > 0 ? `${Math.round((completedExperiments / totalExperiments) * 100)}%` : '100%'}
            </h4>
          </div>
          <RotateCcw className="h-8 w-8 text-violet-500/20" />
        </div>
      </div>

      {/* Visual Analytics Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        
        {/* Bar Chart: Attacks Timeline */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/10 flex flex-col">
          <h2 className="text-base font-semibold tracking-tight mb-6">Recent Injections over Time</h2>
          <div className="h-72 w-full flex-1">
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                  <YAxis stroke="#64748b" fontSize={12} allowDecimals={false} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                  />
                  <Bar dataKey="attacks" fill="#6366f1" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-full flex items-center justify-center text-xs text-slate-500">
                No active metrics yet. Attacks will display once scheduler runs or manual attacks occur.
              </div>
            )}
          </div>
        </div>

        {/* Pie Chart: Attack Type Distribution */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/10 flex flex-col">
          <h2 className="text-base font-semibold tracking-tight mb-6">Attack Type Allocation</h2>
          <div className="h-72 w-full flex-1 flex items-center justify-center">
            {attackData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={attackData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {attackData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                  />
                  <Legend verticalAlign="bottom" height={36} iconType="circle" />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="text-xs text-slate-500">No attack distribution data available yet.</div>
            )}
          </div>
        </div>

        {/* Pie Chart: Container Health */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/10 flex flex-col">
          <h2 className="text-base font-semibold tracking-tight mb-6">Container State Allocation</h2>
          <div className="h-72 w-full flex-1 flex items-center justify-center">
            {containerData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={containerData}
                    cx="50%"
                    cy="50%"
                    innerRadius={0}
                    outerRadius={80}
                    dataKey="value"
                  >
                    {containerData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                  />
                  <Legend verticalAlign="bottom" height={36} iconType="circle" />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="text-xs text-slate-500">No container state data available.</div>
            )}
          </div>
        </div>

        {/* Quick Info list */}
        <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-900/10 flex flex-col justify-between">
          <h2 className="text-base font-semibold tracking-tight mb-4">Scheduler Properties</h2>
          <div className="space-y-4 text-sm flex-1 flex flex-col justify-center">
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span className="text-slate-400">Automation Trigger Mode</span>
              <span className="font-semibold text-violet-400 uppercase">{scheduler?.mode || 'none'}</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span className="text-slate-400">Injection Period Interval</span>
              <span className="font-semibold">{scheduler?.interval || 'N/A'}</span>
            </div>
            <div className="flex justify-between border-b border-slate-800/80 pb-2">
              <span className="text-slate-400">Injection Event Duration</span>
              <span className="font-semibold">{scheduler?.duration || 'N/A'}</span>
            </div>
            <div className="flex justify-between pb-2">
              <span className="text-slate-400">Next Scheduled Attack Cycle</span>
              <span className="font-semibold text-amber-400">
                {scheduler?.next_run ? new Date(scheduler.next_run).toLocaleTimeString() : 'N/A'}
              </span>
            </div>
          </div>
        </div>

      </div>

    </div>
  );
}

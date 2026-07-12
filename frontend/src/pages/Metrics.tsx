import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { 
  Activity, 
  Cpu, 
  Database, 
  Shuffle, 
  AlertTriangle,
  Play
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Legend
} from 'recharts';

interface MetricHistoryPoint {
  time: string;
  totalCpu: number;
  totalMemory: number;
}

export default function Metrics() {
  const [history, setHistory] = useState<MetricHistoryPoint[]>([]);

  // 1. Fetch Real-time container resource values
  const { data: containersRes } = useQuery({
    queryKey: ['containersMetrics'],
    queryFn: () => api.getContainers(),
    refetchInterval: 3000,
  });

  // 2. Fetch raw Prometheus metrics
  const { data: rawMetrics } = useQuery({
    queryKey: ['rawPrometheus'],
    queryFn: () => api.getMetrics(),
    refetchInterval: 3000,
  });

  const containers = containersRes?.data || [];

  // Parse Prometheus metrics text
  const parseMetric = (metricName: string): number => {
    if (!rawMetrics) return 0;
    const regex = new RegExp(`^${metricName}\\s+(\\d+(\\.\\d+)?)`, 'm');
    const match = rawMetrics.match(regex);
    return match ? parseFloat(match[1]) : 0;
  };

  const promMonitored = parseMetric('chaosguard_containers_monitored');
  const promActiveAttacks = parseMetric('chaosguard_active_attacks');
  const promTotalAttacks = parseMetric('chaosguard_attacks_total');

  // Sum CPU and Memory values from active containers
  const totalCpu = parseFloat(containers.reduce((sum, c) => sum + c.cpu_usage, 0).toFixed(1));
  const totalMemoryBytes = containers.reduce((sum, c) => sum + c.memory_usage, 0);
  const totalMemoryMB = parseFloat((totalMemoryBytes / (1024 * 1024)).toFixed(1));

  // Maintain sliding window history
  useEffect(() => {
    const timeLabel = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    setHistory(prev => {
      const next = [...prev, { time: timeLabel, totalCpu, totalMemory: totalMemoryMB }];
      if (next.length > 15) return next.slice(1);
      return next;
    });
  }, [totalCpu, totalMemoryMB]);

  return (
    <div className="space-y-8">
      
      {/* Prometheus Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        
        {/* Active Attacks Card */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Active Attacks</span>
            <h3 className="text-3xl font-extrabold text-amber-500 mt-1">{promActiveAttacks}</h3>
            <p className="text-[9px] text-slate-600 mt-1 font-mono uppercase font-bold">chaosguard_active_attacks</p>
          </div>
          <AlertTriangle className="h-10 w-10 text-amber-500/10" />
        </div>

        {/* Monitored Containers Gauge */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Monitored Targets</span>
            <h3 className="text-3xl font-extrabold text-violet-400 mt-1">{promMonitored}</h3>
            <p className="text-[9px] text-slate-600 mt-1 font-mono uppercase font-bold">chaosguard_containers_monitored</p>
          </div>
          <Activity className="h-10 w-10 text-violet-400/10" />
        </div>

        {/* Total Injections Counter */}
        <div className="p-6 rounded-xl border border-slate-800 bg-[#0f172a]/20 flex items-center justify-between">
          <div>
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Total Injections</span>
            <h3 className="text-3xl font-extrabold text-emerald-400 mt-1">{promTotalAttacks}</h3>
            <p className="text-[9px] text-slate-600 mt-1 font-mono uppercase font-bold">chaosguard_attacks_total</p>
          </div>
          <Play className="h-10 w-10 text-emerald-400/10" />
        </div>

      </div>

      {/* Grid of Resource Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        
        {/* Area Chart: CPU Load Timeline */}
        <div className="p-6 rounded-xl border border-slate-850 bg-slate-900/10 flex flex-col">
          <div className="flex items-center gap-2 mb-6">
            <Cpu className="h-4 w-4 text-indigo-450" />
            <h2 className="text-sm uppercase font-bold tracking-wider">Host CPU utilization (%)</h2>
          </div>
          <div className="h-72 w-full flex-1">
            {history.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={history}>
                  <defs>
                    <linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#6366f1" stopOpacity={0.15}/>
                      <stop offset="95%" stopColor="#6366f1" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="time" stroke="#64748b" fontSize={11} />
                  <YAxis stroke="#64748b" fontSize={11} unit="%" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                  />
                  <Area type="monotone" dataKey="totalCpu" stroke="#6366f1" fillOpacity={1} fill="url(#cpuGrad)" strokeWidth={1.5} />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-full flex items-center justify-center text-xs text-slate-500 font-semibold uppercase tracking-wider">
                Initializing CPU utilization graph...
              </div>
            )}
          </div>
        </div>

        {/* Area Chart: Memory Load Timeline */}
        <div className="p-6 rounded-xl border border-slate-850 bg-slate-900/10 flex flex-col">
          <div className="flex items-center gap-2 mb-6">
            <Database className="h-4 w-4 text-teal-450" />
            <h2 className="text-sm uppercase font-bold tracking-wider">Host Memory utilization (MB)</h2>
          </div>
          <div className="h-72 w-full flex-1">
            {history.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={history}>
                  <defs>
                    <linearGradient id="memGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#0d9488" stopOpacity={0.15}/>
                      <stop offset="95%" stopColor="#0d9488" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="time" stroke="#64748b" fontSize={11} />
                  <YAxis stroke="#64748b" fontSize={11} unit="M" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                  />
                  <Area type="monotone" dataKey="totalMemory" stroke="#0d9488" fillOpacity={1} fill="url(#memGrad)" strokeWidth={1.5} />
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-full flex items-center justify-center text-xs text-slate-500 font-semibold uppercase tracking-wider">
                Initializing Memory utilization graph...
              </div>
            )}
          </div>
        </div>

      </div>

      {/* Bar Chart: Container-by-container allocation */}
      <div className="p-6 rounded-xl border border-slate-850 bg-slate-900/10">
        <div className="flex items-center gap-2 mb-6">
          <Shuffle className="h-4 w-4 text-violet-400" />
          <h2 className="text-sm uppercase font-bold tracking-wider">Container resource allocation overview</h2>
        </div>
        <div className="h-72 w-full">
          {containers.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={containers.map(c => ({
                name: c.name,
                cpu: c.cpu_usage,
                memoryMB: parseFloat((c.memory_usage / (1024 * 1024)).toFixed(1))
              }))}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                <XAxis dataKey="name" stroke="#64748b" fontSize={11} />
                <YAxis stroke="#64748b" fontSize={11} />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', color: '#f8fafc' }}
                />
                <Legend iconType="circle" />
                <Bar dataKey="cpu" name="CPU Usage (%)" fill="#6366f1" radius={[4, 4, 0, 0]} />
                <Bar dataKey="memoryMB" name="Memory (MB)" fill="#0d9488" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-full flex items-center justify-center text-xs text-slate-500">
              No container metrics to plot.
            </div>
          )}
        </div>
      </div>

    </div>
  );
}

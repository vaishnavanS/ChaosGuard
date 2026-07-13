import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../services/api';
import type { Experiment } from '../types';
import { 
  Search, 
  Trash2, 
  Info, 
  X, 
  Calendar,
  AlertOctagon,
  ChevronLeft,
  ChevronRight,
  ShieldCheck,
  Undo2,
  ListFilter
} from 'lucide-react';

export default function Experiments() {
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [attackFilter, setAttackFilter] = useState('all');
  
  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  // Selected experiment for drawer details
  const [selectedExperiment, setSelectedExperiment] = useState<Experiment | null>(null);

  // Queries
  const { data: response, isLoading } = useQuery({
    queryKey: ['experimentsList'],
    queryFn: () => api.getExperiments(),
    refetchInterval: 5000,
  });

  // Mutations
  const deleteExperiment = useMutation({
    mutationFn: (id: string) => api.deleteExperiment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experimentsList'] });
      if (selectedExperiment) {
        setSelectedExperiment(null);
      }
    },
    onError: (err: any) => {
      alert(`Failed to delete record: ${err.message}`);
    }
  });

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-violet-500"></div>
        <p className="text-sm text-gray-400">Loading historical log records...</p>
      </div>
    );
  }

  const list = response?.data || [];

  // Filter list
  const filteredList = list.filter(exp => {
    const matchesSearch = exp.container_name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                          exp.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
                          exp.target_container_id.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' || exp.status === statusFilter;
    const matchesAttack = attackFilter === 'all' || exp.attack_type === attackFilter;

    return matchesSearch && matchesStatus && matchesAttack;
  });

  // Pagination bounds
  const totalPages = Math.ceil(filteredList.length / itemsPerPage);
  const paginatedList = filteredList.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'completed':
      case 'recovered':
        return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
      case 'running':
        return 'bg-amber-500/10 text-amber-400 border border-amber-500/20 animate-pulse';
      case 'pending':
        return 'bg-violet-500/10 text-violet-400 border border-violet-500/20';
      case 'failed':
        return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
      default:
        return 'bg-slate-500/10 text-slate-400 border border-slate-500/20';
    }
  };

  const getSeverityBadge = (attackType: string) => {
    const type = attackType.toLowerCase();
    if (type === 'kill' || type === 'stop') {
      return 'bg-rose-500/10 text-rose-450 border border-rose-500/20 font-bold';
    }
    if (type === 'restart') {
      return 'bg-amber-500/10 text-amber-450 border border-amber-500/20 font-bold';
    }
    return 'bg-sky-500/10 text-sky-455 border border-sky-500/20 font-bold';
  };

  const formatTime = (timeStr: string) => {
    if (!timeStr) return '-';
    try {
      return new Date(timeStr).toLocaleString();
    } catch {
      return timeStr;
    }
  };

  return (
    <div className="relative flex flex-col lg:flex-row gap-6 items-start">
      
      {/* Table Section */}
      <div className="flex-1 w-full space-y-6">
        
        {/* Search and Filters panel */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 items-center justify-between border-b border-slate-800/80 pb-6">
          <div className="flex items-center gap-3 bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2">
            <Search className="h-4 w-4 text-slate-500" />
            <input 
              type="text" 
              placeholder="Search by ID or container..." 
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value);
                setCurrentPage(1);
              }}
              className="bg-transparent text-sm border-0 focus:outline-hidden w-full text-gray-200"
            />
          </div>

          <div className="flex gap-2">
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setCurrentPage(1);
              }}
              className="bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2 text-sm text-gray-300 w-full focus:outline-hidden"
            >
              <option value="all">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="running">Running</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="recovered">Recovered</option>
            </select>
          </div>

          <div className="flex gap-2">
            <select
              value={attackFilter}
              onChange={(e) => {
                setAttackFilter(e.target.value);
                setCurrentPage(1);
              }}
              className="bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2 text-sm text-gray-300 w-full focus:outline-hidden"
            >
              <option value="all">All Attacks</option>
              <option value="pause">Pause</option>
              <option value="stop">Stop</option>
              <option value="restart">Restart</option>
              <option value="kill">Kill</option>
            </select>
          </div>
        </div>

        {/* Paginated Table container */}
        <div className="border border-slate-850 rounded-xl overflow-hidden bg-slate-900/5 shadow-md">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left text-slate-400">
              <thead className="text-[10px] uppercase bg-[#0c101b]/80 text-slate-450 border-b border-slate-800 tracking-wider font-bold">
                <tr>
                  <th className="px-6 py-4">Target Container</th>
                  <th className="px-6 py-4">Attack type</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4">Severity</th>
                  <th className="px-6 py-4">Duration</th>
                  <th className="px-6 py-4">Started At</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-850">
                {paginatedList.length > 0 ? (
                  paginatedList.map((exp) => (
                    <tr 
                      key={exp.id}
                      onClick={() => setSelectedExperiment(exp)}
                      className={`hover:bg-[#0f172a]/20 transition-colors cursor-pointer ${
                        selectedExperiment?.id === exp.id ? 'bg-[#0f172a]/35 border-l-2 border-violet-500' : ''
                      }`}
                    >
                      <td className="px-6 py-4 font-semibold text-gray-200">
                        {exp.container_name || exp.target_container_id.slice(0, 12)}
                      </td>
                      <td className="px-6 py-4">
                        <span className="font-mono text-[10px] uppercase text-violet-400 bg-violet-500/5 px-2 py-0.5 rounded border border-violet-500/10 font-bold">
                          {exp.attack_type}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider ${getStatusBadge(exp.status)}`}>
                          {exp.status}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-0.5 rounded text-[9px] uppercase ${getSeverityBadge(exp.attack_type)}`}>
                          {exp.attack_type === 'kill' || exp.attack_type === 'stop' ? 'Critical' : exp.attack_type === 'restart' ? 'High' : 'Medium'}
                        </span>
                      </td>
                      <td className="px-6 py-4 font-mono text-xs text-gray-300">
                        {exp.duration}s
                      </td>
                      <td className="px-6 py-4 text-xs text-slate-500">
                        {formatTime(exp.started_at)}
                      </td>
                      <td className="px-6 py-4 text-right" onClick={(e) => e.stopPropagation()}>
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => setSelectedExperiment(exp)}
                            className="p-1.5 rounded-lg border border-slate-800 text-slate-450 hover:text-violet-400 cursor-pointer"
                            title="View Parameters"
                          >
                            <Info className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => {
                              if (confirm('Delete this history record?')) {
                                deleteExperiment.mutate(exp.id);
                              }
                            }}
                            className="p-1.5 rounded-lg border border-slate-800 text-slate-450 hover:text-rose-455 cursor-pointer"
                            title="Delete History Record"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center text-slate-500 uppercase tracking-widest text-[10px] font-bold">
                      No experiments have run yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Table pagination stats footer */}
          {totalPages > 1 && (
            <div className="px-6 py-4 flex items-center justify-between border-t border-slate-800 bg-[#0c101b]/40 select-none">
              <span className="text-[10px] text-slate-550 font-bold uppercase tracking-widest">
                Page {currentPage} of {totalPages}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                  disabled={currentPage === 1}
                  className="p-1 rounded-lg border border-slate-800 text-slate-400 hover:bg-slate-850 cursor-pointer disabled:opacity-30 disabled:pointer-events-none"
                >
                  <ChevronLeft className="h-4 w-4" />
                </button>
                <button
                  onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                  disabled={currentPage === totalPages}
                  className="p-1 rounded-lg border border-slate-800 text-slate-400 hover:bg-slate-850 cursor-pointer disabled:opacity-30 disabled:pointer-events-none"
                >
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          )}
        </div>

      </div>

      {/* Details Slide Drawer overlay inside table layout */}
      {selectedExperiment && (
        <div className="w-full lg:w-96 p-6 rounded-xl border border-slate-800 bg-[#0f172a]/95 backdrop-blur-md shadow-2xl flex flex-col gap-4 animate-in slide-in-from-right-10 duration-200">
          <div className="flex justify-between items-start">
            <h3 className="font-bold text-xs uppercase tracking-widest text-slate-400">Disruption Inspector</h3>
            <button 
              onClick={() => setSelectedExperiment(null)}
              className="p-1 rounded-full border border-slate-800 text-slate-450 hover:text-white cursor-pointer"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          <div className="space-y-4 text-xs text-slate-400">
            
            <div className="border-b border-slate-800/80 pb-2">
              <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest">Disruption ID</span>
              <p className="font-mono mt-1 text-gray-300 break-all">{selectedExperiment.id}</p>
            </div>

            <div className="border-b border-slate-800/80 pb-2">
              <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest">Target Container Node</span>
              <p className="font-bold text-sm text-gray-300 mt-1">{selectedExperiment.container_name || '-'}</p>
              <p className="font-mono text-[9px] text-slate-550 break-all mt-0.5">{selectedExperiment.target_container_id}</p>
            </div>

            <div className="grid grid-cols-2 gap-2 border-b border-slate-800/80 pb-2">
              <div>
                <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest">Attack Type</span>
                <p className="font-mono font-bold mt-1 text-violet-400 uppercase">{selectedExperiment.attack_type}</p>
              </div>
              <div>
                <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest">Disruption Duration</span>
                <p className="font-bold text-sm text-gray-300 mt-1">{selectedExperiment.duration} seconds</p>
              </div>
            </div>

            <div className="border-b border-slate-800/80 pb-2">
              <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest">Resilience State</span>
              <div className="mt-1">
                <span className={`px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider ${getStatusBadge(selectedExperiment.status)}`}>
                  {selectedExperiment.status}
                </span>
              </div>
            </div>

            <div className="border-b border-slate-800/80 pb-2">
              <span className="text-[9px] font-bold text-slate-550 uppercase tracking-widest flex items-center gap-1">
                <Calendar className="h-3 w-3" /> Timing Logs
              </span>
              <div className="mt-2 space-y-1 bg-slate-950/20 p-2 rounded">
                <p className="flex justify-between"><span>Started:</span> <span className="font-mono text-gray-300">{formatTime(selectedExperiment.started_at)}</span></p>
                <p className="flex justify-between"><span>Finished:</span> <span className="font-mono text-gray-300">{formatTime(selectedExperiment.ended_at)}</span></p>
              </div>
            </div>

            {selectedExperiment.status === 'recovered' && (
              <div className="border border-emerald-950 bg-emerald-950/15 p-3 rounded-lg text-emerald-450 flex items-center gap-2">
                <ShieldCheck className="h-4 w-4" />
                <div className="flex flex-col">
                  <span className="uppercase text-[9px] font-bold">Auto-Recovery SLA Active</span>
                  <span className="text-[10px] text-slate-400 mt-0.5">Container automatically self-healed by ChaosGuard Scheduler.</span>
                </div>
              </div>
            )}

            {selectedExperiment.error_message && (
              <div className="border border-rose-950/40 bg-rose-950/5 p-3 rounded-lg text-rose-400">
                <span className="uppercase text-[9px] font-bold flex items-center gap-1">
                  <AlertOctagon className="h-3.5 w-3.5" /> Disruption Failure Error
                </span>
                <p className="mt-1 text-xs text-rose-350/80 leading-relaxed font-mono">{selectedExperiment.error_message}</p>
              </div>
            )}

            <div className="border-b border-slate-800/80 pb-2">
              <span className="text-[9px] font-bold text-slate-555 uppercase tracking-widest">Injected Parameters</span>
              <pre className="font-mono text-[10px] bg-[#0c1222] p-3 rounded-lg border border-slate-800 text-gray-300 mt-1 overflow-x-auto">
                {JSON.stringify(JSON.parse(selectedExperiment.parameters || '{}'), null, 2)}
              </pre>
            </div>

          </div>
        </div>
      )}

    </div>
  );
}

import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { 
  Terminal, 
  Search, 
  Download, 
  RefreshCcw,
  Play,
  Pause
} from 'lucide-react';

export default function Logs() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLevel, setSelectedLevel] = useState<'ALL' | 'INF' | 'DBG' | 'WRN' | 'ERR'>('ALL');
  const [autoScroll, setAutoScroll] = useState(true);

  // Queries
  const { data: logsRes, isLoading, refetch } = useQuery({
    queryKey: ['logsData'],
    queryFn: () => api.getLogs(),
    refetchInterval: 3000,
  });

  const logs = logsRes?.data || [];

  // Filter logs locally
  const filteredLogs = logs.filter(line => {
    const matchesSearch = line.toLowerCase().includes(searchTerm.toLowerCase());
    if (selectedLevel === 'ALL') return matchesSearch;
    
    // Parse level from Zerolog format, e.g. "INF", "DBG", "WRN", "ERR"
    const matchesLevel = line.includes(`"level":"${selectedLevel.toLowerCase()}"`) || 
                         line.includes(`"level":"${selectedLevel.toUpperCase()}"`) ||
                         line.includes(` ${selectedLevel} `) ||
                         line.includes(`INF`) && selectedLevel === 'INF' ||
                         line.includes(`DBG`) && selectedLevel === 'DBG' ||
                         line.includes(`WRN`) && selectedLevel === 'WRN' ||
                         line.includes(`ERR`) && selectedLevel === 'ERR';
    return matchesSearch && matchesLevel;
  });

  // Scroll to bottom when new logs arrive if autoScroll is active
  useEffect(() => {
    if (autoScroll) {
      const el = document.getElementById('log-terminal');
      if (el) el.scrollTop = el.scrollHeight;
    }
  }, [filteredLogs, autoScroll]);

  // Clean log message helper (strips zerolog JSON keys for terminal rendering)
  const formatLogLine = (line: string): string => {
    try {
      const parsed = JSON.parse(line);
      const time = parsed.time ? new Date(parsed.time).toLocaleTimeString() : '';
      const level = (parsed.level || 'INF').toUpperCase();
      const msg = parsed.message || parsed.msg || '';
      return `[${time}] [${level}] ${msg}`;
    } catch {
      return line;
    }
  };

  const getLineClass = (line: string): string => {
    const lower = line.toLowerCase();
    if (lower.includes('"level":"err"') || lower.includes('"level":"error"') || lower.includes('error')) {
      return 'text-rose-450';
    }
    if (lower.includes('"level":"wrn"') || lower.includes('"level":"warn"') || lower.includes('warning')) {
      return 'text-amber-450';
    }
    if (lower.includes('"level":"dbg"') || lower.includes('"level":"debug"') || lower.includes('debug')) {
      return 'text-slate-500';
    }
    return 'text-sky-450';
  };

  const handleDownloadLogs = () => {
    const blob = new Blob([logs.join('\n')], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `chaosguard-${new Date().toISOString().slice(0, 10)}.log`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="space-y-6">
      
      {/* Logs Filters toolbar */}
      <div className="flex flex-col sm:flex-row gap-4 items-stretch sm:items-center justify-between border-b border-slate-800/80 pb-6">
        
        {/* Search */}
        <div className="flex items-center gap-3 bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2 w-full sm:w-80">
          <Search className="h-4 w-4 text-slate-500" />
          <input 
            type="text" 
            placeholder="Search matching logs..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="bg-transparent text-sm border-0 focus:outline-hidden w-full text-gray-200"
          />
        </div>

        {/* Level Filters and Action Controls */}
        <div className="flex flex-wrap gap-2 items-center">
          
          {/* Level Filter Selector */}
          <div className="flex border border-slate-800 rounded-lg overflow-hidden">
            {(['ALL', 'DBG', 'INF', 'WRN', 'ERR'] as const).map((level) => (
              <button
                key={level}
                onClick={() => setSelectedLevel(level)}
                className={`px-3 py-1.5 text-xs font-semibold uppercase tracking-wider cursor-pointer border-r border-slate-800/80 last:border-0 ${
                  selectedLevel === level 
                    ? 'bg-violet-600/15 text-violet-400 font-bold' 
                    : 'text-slate-400 hover:bg-slate-850'
                }`}
              >
                {level}
              </button>
            ))}
          </div>

          {/* Pause Scroll Toggle */}
          <button
            onClick={() => setAutoScroll(prev => !prev)}
            className={`p-2 rounded-lg border cursor-pointer ${
              autoScroll 
                ? 'border-slate-800 text-slate-400 hover:bg-slate-850' 
                : 'border-amber-500/20 bg-amber-500/10 text-amber-400'
            }`}
            title={autoScroll ? 'Pause Live Autoscroll' : 'Resume Live Autoscroll'}
          >
            {autoScroll ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4 animate-pulse" />}
          </button>

          {/* Refresh Logs list */}
          <button
            onClick={() => refetch()}
            className="p-2 border border-slate-800 rounded-lg text-slate-400 hover:bg-slate-850 cursor-pointer"
            title="Force Fetch Logs"
          >
            <RefreshCcw className="h-4 w-4" />
          </button>

          {/* Download Logs */}
          <button
            onClick={handleDownloadLogs}
            disabled={logs.length === 0}
            className="p-2 border border-slate-800 rounded-lg text-slate-400 hover:bg-slate-850 cursor-pointer disabled:opacity-30 disabled:pointer-events-none"
            title="Download Logs File"
          >
            <Download className="h-4 w-4" />
          </button>

        </div>
      </div>

      {/* Terminal logs viewer */}
      <div className="relative rounded-xl border border-slate-850 bg-[#070b13] shadow-2xl flex flex-col h-[70vh]">
        
        {/* Terminal Header */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-850/80 bg-[#0c1222]/80 shrink-0 text-slate-400 select-none">
          <Terminal className="h-4 w-4 text-violet-500" />
          <span className="text-[10px] uppercase font-bold tracking-widest text-slate-500">Live Host Terminal Logs</span>
        </div>

        {/* Scroll Box */}
        <div 
          id="log-terminal"
          className="flex-1 overflow-auto p-6 font-mono text-xs leading-relaxed space-y-1.5 scrollbar-thin select-text"
        >
          {isLoading ? (
            <div className="h-full flex items-center justify-center text-slate-500 animate-pulse uppercase tracking-wider text-[10px] font-bold">
              Connecting terminal socket...
            </div>
          ) : filteredLogs.length > 0 ? (
            filteredLogs.map((line, index) => (
              <div 
                key={index}
                className={`py-0.5 px-2 hover:bg-slate-900/50 rounded transition-colors whitespace-pre-wrap break-all ${getLineClass(line)}`}
              >
                {formatLogLine(line)}
              </div>
            ))
          ) : (
            <div className="h-full flex items-center justify-center text-slate-600">
              No matching log lines recorded.
            </div>
          )}
        </div>

      </div>

    </div>
  );
}

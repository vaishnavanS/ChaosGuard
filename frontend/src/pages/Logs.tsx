import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { 
  Terminal, 
  Search, 
  Download, 
  RefreshCcw
} from 'lucide-react';

export default function Logs() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLevel, setSelectedLevel] = useState<'ALL' | 'INF' | 'DBG' | 'WRN' | 'ERR'>('ALL');
  const [autoScroll, setAutoScroll] = useState(true);

  // Queries
  const { data: logsRes, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['liveLogs'],
    queryFn: () => api.getLogs(),
    refetchInterval: 3000, // Poll logs every 3 seconds
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

  // Clean log message helper (strips zerolog JSON keys for terminal rendering if raw is too ugly)
  const formatLogLine = (line: string): string => {
    try {
      const parsed = JSON.parse(line);
      const time = parsed.time ? new Date(parsed.time).toLocaleTimeString() : '';
      const level = (parsed.level || 'INF').toUpperCase();
      const msg = parsed.message || parsed.msg || '';
      


      return `[${time}] [${level}] ${msg}`;
    } catch {
      // Fallback to raw line rendering if not JSON format
      return line.trim();
    }
  };

  const getLineClass = (line: string): string => {
    if (line.includes('"level":"err"') || line.includes('"level":"error"') || line.includes('ERR')) {
      return 'text-rose-400';
    }
    if (line.includes('"level":"warn"') || line.includes('WRN')) {
      return 'text-amber-400';
    }
    if (line.includes('"level":"debug"') || line.includes('DBG')) {
      return 'text-slate-500';
    }
    return 'text-slate-300';
  };

  const downloadLogsFile = () => {
    const content = logs.join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `chaosguard-daemon-${new Date().toISOString().slice(0,10)}.log`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6 flex flex-col h-[calc(100vh-9rem)]">
      
      {/* Search and Filters toolbar */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 items-center justify-between border-b border-slate-800/80 pb-6 shrink-0">
        
        {/* Search */}
        <div className="flex items-center gap-3 bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2">
          <Search className="h-4 w-4 text-slate-500" />
          <input 
            type="text" 
            placeholder="Search log logs..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="bg-transparent text-sm border-0 focus:outline-hidden w-full text-gray-200"
          />
        </div>

        {/* Level Filter */}
        <div className="flex gap-2">
          <select
            value={selectedLevel}
            onChange={(e) => setSelectedLevel(e.target.value as any)}
            className="bg-slate-900/40 border border-slate-800 rounded-lg px-3 py-2 text-sm text-gray-300 w-full focus:outline-hidden"
          >
            <option value="ALL">All Levels</option>
            <option value="INF">Info</option>
            <option value="DBG">Debug</option>
            <option value="WRN">Warning</option>
            <option value="ERR">Error</option>
          </select>
        </div>

        {/* Auto Scroll Checkbox */}
        <label className="flex items-center gap-2 text-xs font-semibold text-slate-400 select-none cursor-pointer">
          <input 
            type="checkbox" 
            checked={autoScroll} 
            onChange={(e) => setAutoScroll(e.target.checked)}
            className="rounded border-slate-800 bg-slate-900 text-violet-600 focus:ring-violet-500 focus:ring-offset-0 focus:ring-opacity-0 h-4 w-4"
          />
          <span>AUTO SCROLL TO BOTTOM</span>
        </label>

        {/* Action buttons */}
        <div className="flex justify-end gap-2">
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="p-2 rounded-lg border border-slate-800 text-slate-400 hover:text-violet-400 hover:bg-slate-800/40 cursor-pointer disabled:opacity-50"
            title="Refresh Logs Buffer"
          >
            <RefreshCcw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={downloadLogsFile}
            className="p-2 rounded-lg border border-slate-800 text-slate-400 hover:text-emerald-400 hover:bg-slate-800/40 cursor-pointer flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider px-3"
            title="Download log logs"
          >
            <Download className="h-4 w-4" />
            <span>Download</span>
          </button>
        </div>

      </div>

      {/* Terminal View area */}
      <div className="flex-1 min-h-0 border border-slate-850 rounded-xl overflow-hidden bg-[#070b13] flex flex-col font-mono">
        <div className="px-4 py-2 bg-slate-950/80 border-b border-slate-900/60 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-violet-500" />
            <span className="text-xs text-slate-500 font-bold tracking-wider">DAEMON STDOUT TERMINAL</span>
          </div>
          <span className="text-[10px] text-slate-600 font-mono">Buffer size: {filteredLogs.length} lines</span>
        </div>

        <div 
          id="log-terminal"
          className="flex-1 overflow-y-auto p-4 space-y-1.5 text-xs select-text scroll-smooth"
        >
          {isLoading ? (
            <div className="h-full flex flex-col items-center justify-center gap-2 text-slate-600 font-sans">
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-slate-800"></div>
              <span>Fetching logs from daemon...</span>
            </div>
          ) : filteredLogs.length > 0 ? (
            filteredLogs.map((line, i) => (
              <div key={i} className={`break-all leading-relaxed whitespace-pre-wrap ${getLineClass(line)}`}>
                <span className="text-slate-700 select-none mr-3 inline-block w-8 text-right">{(i + 1).toString().padStart(3, '0')}</span>
                <span>{formatLogLine(line)}</span>
              </div>
            ))
          ) : (
            <div className="h-full flex items-center justify-center text-slate-600 font-sans text-xs">
              No matching logs inside buffer.
            </div>
          )}
        </div>
      </div>

    </div>
  );
}

// Success API wrapper structure
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}

// Health Check payload schema
export interface HealthResponse {
  status: string;
  state: string;
  version: string;
}

// Docker Container details schema
export interface Container {
  id: string;
  name: string;
  image: string;
  status: string; // e.g. "running", "paused", "exited"
  state: string;
  is_monitored: boolean;
  uptime: number; // in seconds
  cpu_usage: number; // in percent
  memory_usage: number; // in bytes
  labels: Record<string, string>;
}

// Chaos Experiment record schema
export interface Experiment {
  id: string;
  target_container_id: string;
  container_name: string;
  attack_type: string; // e.g. "pause", "stop", "restart", "kill"
  duration: number; // in seconds
  status: string; // e.g. "pending", "running", "completed", "failed", "recovered"
  parameters: string; // JSON configuration string
  started_at: string;
  ended_at: string;
  error_message: string;
}

// Scheduler state properties schema
export interface SchedulerStatus {
  running: boolean;
  mode: string; // e.g. "random", "round-robin"
  interval: string;
  duration: string;
  next_run: string;
}

// Runtime lifecycle state properties schema
export interface RuntimeResponse {
  state: string; // e.g. "running", "starting", "stopping", "stopped"
}

// Stats metrics card overview definitions
export interface StatOverview {
  runtimeState: string;
  schedulerRunning: boolean;
  totalContainers: number;
  runningContainers: number;
  activeExperiments: number;
  totalExperiments: number;
  successfulRecoveries: number;
  failedRecoveries: number;
  totalAttacks: number;
}

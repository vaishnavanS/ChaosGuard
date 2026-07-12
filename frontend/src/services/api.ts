import axios from 'axios';
import type { ApiResponse, HealthResponse, Container, Experiment, SchedulerStatus, RuntimeResponse } from '../types';

// Reads dynamic backend address from local storage, fallback to standard localhost
export const getBackendURL = (): string => {
  return localStorage.getItem('chaosguard_backend_url') || 'http://localhost:8080';
};

// Creates dynamic Axios instance
const client = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to apply current configured backend URL
client.interceptors.request.use((config) => {
  config.baseURL = getBackendURL();
  return config;
});

export const api = {
  // System Health
  async getHealth(): Promise<ApiResponse<HealthResponse>> {
    const res = await client.get<ApiResponse<HealthResponse>>('/health');
    return res.data;
  },

  // Monitored Container Controls
  async getContainers(): Promise<ApiResponse<Container[]>> {
    const res = await client.get<ApiResponse<Container[]>>('/containers');
    return res.data;
  },

  async getContainer(id: string): Promise<ApiResponse<Container>> {
    const res = await client.get<ApiResponse<Container>>(`/containers/${id}`);
    return res.data;
  },

  // Chaos Experiments Operations
  async getExperiments(): Promise<ApiResponse<Experiment[]>> {
    const res = await client.get<ApiResponse<Experiment[]>>('/experiments');
    return res.data;
  },

  async getExperiment(id: string): Promise<ApiResponse<Experiment>> {
    const res = await client.get<ApiResponse<Experiment>>(`/experiments/${id}`);
    return res.data;
  },

  async createExperiment(targetContainerId: string, attackType: string, durationSec: number): Promise<ApiResponse<Experiment>> {
    const res = await client.post<ApiResponse<Experiment>>('/experiments', {
      target_container_id: targetContainerId,
      attack_type: attackType,
      duration: durationSec,
    });
    return res.data;
  },

  async deleteExperiment(id: string): Promise<ApiResponse<void>> {
    const res = await client.delete<ApiResponse<void>>(`/experiments/${id}`);
    return res.data;
  },

  // Automated Scheduler Controls
  async getSchedulerStatus(): Promise<ApiResponse<SchedulerStatus>> {
    const res = await client.get<ApiResponse<SchedulerStatus>>('/scheduler/status');
    return res.data;
  },

  async startScheduler(): Promise<ApiResponse<void>> {
    const res = await client.post<ApiResponse<void>>('/scheduler/start');
    return res.data;
  },

  async stopScheduler(): Promise<ApiResponse<void>> {
    const res = await client.post<ApiResponse<void>>('/scheduler/stop');
    return res.data;
  },

  // Daemon Runtime Controls
  async getRuntime(): Promise<ApiResponse<RuntimeResponse>> {
    const res = await client.get<ApiResponse<RuntimeResponse>>('/runtime');
    return res.data;
  },

  async stopRuntime(): Promise<ApiResponse<void>> {
    const res = await client.post<ApiResponse<void>>('/runtime/stop');
    return res.data;
  },

  // Live Logs Buffer
  async getLogs(): Promise<ApiResponse<string[]>> {
    const res = await client.get<ApiResponse<string[]>>('/logs');
    return res.data;
  },

  // Raw Prometheus metrics fetching
  async getMetrics(): Promise<string> {
    const res = await client.get<string>('/metrics', { responseType: 'text' });
    return res.data;
  },
};

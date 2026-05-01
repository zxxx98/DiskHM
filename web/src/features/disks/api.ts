import { http } from '../../lib/http';

export type DiskListItem = {
  id: string;
  name: string;
  path: string;
  model: string;
  mounts?: string[];
  powerState: string;
  refreshFreshness: string;
  task?: {
    kind: string;
    state: string;
    executeAt?: string;
    lastError?: string;
  };
  unsupported: boolean;
};

export function fetchDisks() {
  return http.json<{ items: DiskListItem[] }>('/api/disks');
}

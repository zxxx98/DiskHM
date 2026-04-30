import { http } from '../../lib/http';

export type DiskListItem = {
  id: string;
  name: string;
  model: string;
  powerState: string;
  refreshFreshness: string;
  unsupported: boolean;
};

export function fetchDisks() {
  return http.json<{ items: DiskListItem[] }>('/api/disks');
}

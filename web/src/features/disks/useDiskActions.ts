import { useQueryClient } from '@tanstack/react-query';

import { http } from '../../lib/http';

export function useDiskActions() {
  const queryClient = useQueryClient();

  async function postAction(path: string, body?: object) {
    const response = await http.request(path, {
      method: 'POST',
      body,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    await queryClient.invalidateQueries({ queryKey: ['disks'] });
    await queryClient.invalidateQueries({ queryKey: ['topology'] });
    await queryClient.invalidateQueries({ queryKey: ['events'] });
  }

  return {
    sleepNow: (diskID: string) => postAction(`/api/disks/${diskID}/sleep-now`),
    sleepAfter: (diskID: string, minutes: number) => postAction(`/api/disks/${diskID}/sleep-after`, { minutes }),
    refreshWake: (diskID: string) => postAction(`/api/disks/${diskID}/refresh-wake`),
  };
}

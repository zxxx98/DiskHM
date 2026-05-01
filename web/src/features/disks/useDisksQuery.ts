import { useQuery } from '@tanstack/react-query';

import { fetchDisks } from './api';

export function useDisksQuery() {
  return useQuery({
    queryKey: ['disks'],
    queryFn: async () => {
      const payload = await fetchDisks();
      return payload.items;
    },
    retry: false,
  });
}

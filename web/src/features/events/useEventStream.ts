import { useEffect } from 'react';

export function useEventStream(onMessage: (event: MessageEvent<string>) => void) {
  useEffect(() => {
    const source = new EventSource('/api/events/stream', { withCredentials: true });
    source.onmessage = onMessage;

    return () => {
      source.close();
    };
  }, [onMessage]);
}

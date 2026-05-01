import { useState, useEffectEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { RouteObject } from 'react-router-dom';

import { DiskTablePage } from '../features/disks/DiskTablePage';
import { useDiskActions } from '../features/disks/useDiskActions';
import { useDisksQuery } from '../features/disks/useDisksQuery';
import { EventsPage } from '../features/events/EventsPage';
import { useEventStream } from '../features/events/useEventStream';
import { SettingsPage } from '../features/settings/SettingsPage';
import { LoginPage } from '../features/session/LoginPage';
import { TopologyPage } from '../features/topology/TopologyPage';
import { http } from '../lib/http';

type TopologyNode = {
  id: string;
  label: string;
};

type TopologyEdge = {
  from: string;
  to: string;
};

type TopologyResponse = {
  edges: TopologyEdge[];
  nodes: TopologyNode[];
};

type SettingsResponse = {
  quiet_grace_seconds: number;
};

type EventItem = {
  id: string;
  message: string;
};

type EventsResponse = {
  items: Array<{
    diskId: string;
    id: number;
    kind: string;
    message: string;
  }>;
};

function DisksRoute() {
  const { data, error, isLoading } = useDisksQuery();
  const { refreshWake, sleepAfter, sleepNow } = useDiskActions();
  const [actionError, setActionError] = useState<string | null>(null);

  async function runAction(action: () => Promise<void>) {
    try {
      setActionError(null);
      await action();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : 'Disk action failed.');
    }
  }

  if (isLoading) {
    return (
      <DiskTablePage
        actionError={null}
        disks={[]}
      />
    );
  }

  if (error) {
    return (
      <DiskTablePage
        actionError={error instanceof Error ? error.message : 'Failed to load disks.'}
        disks={[]}
      />
    );
  }

  return (
    <DiskTablePage
      actionError={actionError}
      disks={data ?? []}
      onSleepLater={(diskID) => void runAction(() => sleepAfter(diskID, 30))}
      onSleepNow={(diskID) => void runAction(() => sleepNow(diskID))}
      onWakeRefresh={(diskID) => void runAction(() => refreshWake(diskID))}
    />
  );
}

function TopologyRoute() {
  const { data } = useQuery({
    queryKey: ['topology'],
    queryFn: () => http.json<TopologyResponse>('/api/topology'),
  });

  return <TopologyPage edges={data?.edges ?? []} nodes={data?.nodes ?? []} />;
}

function EventsRoute() {
  const { data } = useQuery({
    queryKey: ['events'],
    queryFn: () => http.json<EventsResponse>('/api/events'),
  });
  const [streamedEvents, setStreamedEvents] = useState<EventItem[]>([]);
  const onMessage = useEffectEvent((event: MessageEvent<string>) => {
    const payload = JSON.parse(event.data) as { id: number; message: string };
    setStreamedEvents((current) => [{ id: `stream-${payload.id}`, message: payload.message }, ...current].slice(0, 25));
  });

  useEventStream(onMessage);

  const events: EventItem[] = [
    ...(streamedEvents ?? []),
    ...((data?.items ?? []).map((event) => ({
      id: `history-${event.id}`,
      message: event.message,
    })) as EventItem[]),
  ];

  return <EventsPage events={events} />;
}

function SettingsRoute() {
  const { data } = useQuery({
    queryKey: ['settings'],
    queryFn: () => http.json<SettingsResponse>('/api/settings'),
  });

  return <SettingsPage quietGraceSeconds={data?.quiet_grace_seconds ?? 10} />;
}

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <LoginPage />,
  },
  {
    path: '/disks',
    element: <DisksRoute />,
  },
  {
    path: '/topology',
    element: <TopologyRoute />,
  },
  {
    path: '/events',
    element: <EventsRoute />,
  },
  {
    path: '/settings',
    element: <SettingsRoute />,
  },
];

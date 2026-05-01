import { useState, useEffectEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { RouteObject } from 'react-router-dom';

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

function TopologyRoute() {
  const { data } = useQuery({
    queryKey: ['topology'],
    queryFn: () => http.json<TopologyResponse>('/api/topology'),
  });

  return <TopologyPage edges={data?.edges ?? []} nodes={data?.nodes ?? []} />;
}

function EventsRoute() {
  const [events, setEvents] = useState<EventItem[]>([]);
  const onMessage = useEffectEvent((event: MessageEvent<string>) => {
    setEvents((current) => [{ id: `${Date.now()}-${current.length}`, message: event.data }, ...current].slice(0, 25));
  });

  useEventStream(onMessage);

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

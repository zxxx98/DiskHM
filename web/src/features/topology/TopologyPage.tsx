type TopologyNode = {
  id: string;
  label: string;
};

type TopologyEdge = {
  from: string;
  to: string;
};

type TopologyPageProps = {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
};

export function TopologyPage({ nodes, edges }: TopologyPageProps) {
  const edgeLabel = `${edges.length} edge${edges.length === 1 ? '' : 's'}`;

  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="topology-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">System map</p>
            <h1 id="topology-page-title">Topology</h1>
          </div>
          <p className="disk-page__copy">{edgeLabel}</p>
        </div>

        <ul aria-label="Topology nodes">
          {nodes.length === 0 ? <li>No topology nodes reported yet.</li> : null}
          {nodes.map((node) => (
            <li key={node.id}>{node.label}</li>
          ))}
        </ul>
      </section>
    </main>
  );
}

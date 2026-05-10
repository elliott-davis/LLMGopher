'use client';

import type { RoutePolicy, RouteTarget } from '@/lib/admin-surface-contracts';

interface RouteDiagramProps {
  route: RoutePolicy;
}

export function RouteDiagram({ route }: RouteDiagramProps) {
  return (
    <div role="img" aria-label={`${route.strategy} routing strategy for ${route.model_alias}`}>
      {route.strategy === 'single' && <SingleDiagram targets={route.targets} />}
      {route.strategy === 'fallback' && <FallbackDiagram targets={route.targets} />}
      {route.strategy === 'weighted' && <WeightedDiagram targets={route.targets} />}
      {route.strategy === 'latency' && <LatencyDiagram targets={route.targets} />}
    </div>
  );
}

function SingleDiagram({ targets }: { targets: RouteTarget[] }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <span>→</span>
      <span>{targets[0]?.provider_name ?? 'No provider'}</span>
    </div>
  );
}

function FallbackDiagram({ targets }: { targets: RouteTarget[] }) {
  return (
    <ol className="text-sm space-y-1">
      {targets.map((t, i) => (
        <li key={t.provider_id} className="flex items-center gap-2">
          <span className="text-muted-foreground">{i === 0 ? 'Primary' : `Fallback ${i}`}:</span>
          <span data-flow-strip={i === 0 ? 'primary' : 'secondary'}>{t.provider_name}</span>
        </li>
      ))}
    </ol>
  );
}

function WeightedDiagram({ targets }: { targets: RouteTarget[] }) {
  const total = targets.reduce((sum, t) => sum + t.weight, 0);
  return (
    <ol className="text-sm space-y-1">
      {targets.map((t) => (
        <li key={t.provider_id} className="flex items-center gap-2">
          <span>{t.provider_name}:</span>
          <span>{total > 0 ? Math.round((t.weight / total) * 100) : 0}%</span>
          <div
            data-testid="flow-curve-primary"
            style={{ width: `${total > 0 ? (t.weight / total) * 100 : 0}%`, height: 8 }}
            className="bg-primary rounded"
          />
        </li>
      ))}
    </ol>
  );
}

function LatencyDiagram({ targets }: { targets: RouteTarget[] }) {
  return (
    <ol className="text-sm space-y-1">
      {targets.map((t) => (
        <li key={t.provider_id} className="flex items-center gap-2">
          <span>{t.provider_name}:</span>
          <span>{t.latency_ms != null ? `${t.latency_ms}ms` : '—'}</span>
        </li>
      ))}
    </ol>
  );
}

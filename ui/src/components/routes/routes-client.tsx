'use client';

import { useState } from 'react';
import type { RoutePolicy } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';

interface RoutesClientProps {
  routes: RoutePolicy[];
  unavailable: boolean;
}

const STRATEGIES = ['all', 'single', 'fallback', 'weighted', 'latency'] as const;
type StrategyFilter = typeof STRATEGIES[number];

export function RoutesClient({ routes, unavailable }: RoutesClientProps) {
  const [selectedStrategy, setSelectedStrategy] = useState<StrategyFilter>('all');
  const [selectedRoute, setSelectedRoute] = useState<RoutePolicy | null>(null);

  const filtered = selectedStrategy === 'all' ? routes : routes.filter(r => r.strategy === selectedStrategy);

  return (
    <>
      <PageHeader title="Routes" />
      <div className="content-area">
        <div className="flex gap-2 mb-4">
          <button
            data-testid="strategy-switcher"
            className="border rounded px-3 py-1 text-sm"
            onClick={() => {}}
          >
            Filter Strategy
          </button>
          {STRATEGIES.slice(1).map((s) => (
            <button
              key={s}
              data-testid={`strategy-${s}`}
              onClick={() => setSelectedStrategy(s)}
              className={`px-3 py-1 rounded text-sm border ${selectedStrategy === s ? 'bg-primary text-primary-foreground border-primary' : 'border-border'}`}
            >
              {s}
            </button>
          ))}
        </div>

        <div className="mb-4 p-3 border rounded bg-amber-50 text-amber-800 text-sm" data-testid="route-save-unavailable">
          Route mutation controls are not yet available. Production writes will be enabled when the admin API contract is reconciled.
        </div>

        {unavailable && <UnavailableState message="Routes are not yet available in this environment." />}
        {!unavailable && filtered.length === 0 && <EmptyState message="No routes match the current filter." />}
        {!unavailable && filtered.length > 0 && (
          <div className="space-y-2">
            {filtered.map((route) => (
              <RouteRow key={route.id} route={route} onClick={setSelectedRoute} />
            ))}
          </div>
        )}

        {selectedRoute && (
          <RouteDetailPanel route={selectedRoute} onClose={() => setSelectedRoute(null)} />
        )}
      </div>
    </>
  );
}

function RouteRow({ route, onClick }: { route: RoutePolicy; onClick: (r: RoutePolicy) => void }) {
  return (
    <div
      data-testid={`route-row-${route.id}`}
      onClick={() => onClick(route)}
      className="border rounded p-3 cursor-pointer hover:bg-muted/50 flex items-center justify-between"
    >
      <div>
        <span className="font-medium text-sm">{route.model_alias}</span>
        <span
          className="ml-2 text-xs text-muted-foreground capitalize"
          data-testid={`strategy-view-${route.strategy}`}
        >
          {route.strategy}
        </span>
      </div>
      <span className={`text-xs px-2 py-0.5 rounded ${route.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
        {route.enabled ? 'Enabled' : 'Disabled'}
      </span>
    </div>
  );
}

function RouteDetailPanel({ route, onClose }: { route: RoutePolicy; onClose: () => void }) {
  return (
    <aside data-testid="route-detail-panel" className="fixed inset-y-0 right-0 w-[480px] bg-background border-l shadow-xl flex flex-col z-50 p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-medium">{route.model_alias}</h2>
        <button onClick={onClose} aria-label="Close">✕</button>
      </div>
      <div className="space-y-2">
        <div data-testid={`strategy-view-${route.strategy}`}>
          <p className="text-sm text-muted-foreground capitalize">Strategy: {route.strategy}</p>
        </div>
        <div className="space-y-1">
          {route.targets.map((target, i) => (
            <div key={target.provider_id} className="text-sm border rounded p-2">
              <span className="font-medium">{target.provider_name}</span>
              {route.strategy === 'fallback' && (
                <span className="text-xs ml-2 text-muted-foreground">
                  {i === 0 ? 'Primary' : `Fallback ${i}`}
                </span>
              )}
              {route.strategy === 'weighted' && (
                <span className="text-xs ml-2 text-muted-foreground">{target.weight}%</span>
              )}
            </div>
          ))}
        </div>
      </div>
    </aside>
  );
}

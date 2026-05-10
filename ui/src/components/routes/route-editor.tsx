'use client';

import type { RoutePolicy } from '@/lib/admin-surface-contracts';
import { validateRoutePolicy } from './route-validation';

interface RouteEditorProps {
  route: RoutePolicy;
}

export function RouteEditor({ route }: RouteEditorProps) {
  const errors = validateRoutePolicy(route);
  return (
    <div>
      {errors.length > 0 && (
        <div role="alert" className="text-sm text-red-600 space-y-1 mb-3">
          {errors.map((e) => <p key={e}>{e}</p>)}
        </div>
      )}
      <div className="space-y-2">
        {route.targets.map((t) => (
          <div key={t.provider_id} className="border rounded p-2 text-sm flex items-center justify-between">
            <span>{t.provider_name}</span>
            {route.strategy === 'weighted' && (
              <input
                data-testid="weight-slider"
                type="number"
                defaultValue={t.weight}
                min={0}
                className="border rounded w-20 px-2 py-1 text-xs"
              />
            )}
          </div>
        ))}
      </div>
      <div className="mt-4 p-3 border rounded bg-amber-50 text-amber-800 text-xs">
        Production saves are unavailable until the route mutation API contract is reconciled.
      </div>
    </div>
  );
}

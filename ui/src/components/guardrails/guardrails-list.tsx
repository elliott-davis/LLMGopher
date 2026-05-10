'use client';

import type { GuardrailRule } from '@/lib/admin-surface-contracts';

interface GuardrailsListProps {
  guardrails: GuardrailRule[];
  onToggle: (id: string, enabled: boolean) => void;
}

export function GuardrailsList({ guardrails, onToggle }: GuardrailsListProps) {
  return (
    <div className="space-y-2">
      {guardrails.map((g) => (
        <div key={g.id} data-testid={`guardrail-row-${g.id}`} className="border rounded p-3 flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">{g.display_name}</p>
            {g.description && <p className="text-xs text-muted-foreground">{g.description}</p>}
          </div>
          <button
            data-testid={`guardrail-toggle-${g.id}`}
            role="switch"
            aria-checked={g.enabled}
            onClick={() => onToggle(g.id, !g.enabled)}
            className={`relative inline-flex h-5 w-10 items-center rounded-full transition-colors ${g.enabled ? 'bg-primary' : 'bg-muted'}`}
          >
            <span className={`inline-block h-3 w-3 transform rounded-full bg-white transition-transform ${g.enabled ? 'translate-x-5' : 'translate-x-1'}`} />
          </button>
        </div>
      ))}
    </div>
  );
}

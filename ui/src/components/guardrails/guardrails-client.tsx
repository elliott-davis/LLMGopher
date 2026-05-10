'use client';

import { useState } from 'react';
import type { GuardrailRule } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';
import { GuardrailsList } from './guardrails-list';

interface GuardrailsClientProps {
  guardrails: GuardrailRule[];
  unavailable: boolean;
}

export function GuardrailsClient({ guardrails: initial, unavailable }: GuardrailsClientProps) {
  const [guardrails, setGuardrails] = useState(initial);

  const handleToggle = async (id: string, enabled: boolean) => {
    setGuardrails((prev) => prev.map((g) => g.id === id ? { ...g, enabled } : g));
    try {
      const res = await fetch(`/api/guardrails/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      });
      if (!res.ok) {
        setGuardrails((prev) => prev.map((g) => g.id === id ? { ...g, enabled: !enabled } : g));
      }
    } catch {
      setGuardrails((prev) => prev.map((g) => g.id === id ? { ...g, enabled: !enabled } : g));
    }
  };

  return (
    <>
      <PageHeader title="Guardrails" />
      <div className="content-area">
        {unavailable && <UnavailableState message="Guardrails are not yet available in this environment." />}
        {!unavailable && guardrails.length === 0 && <EmptyState message="No guardrail rules configured." />}
        {!unavailable && guardrails.length > 0 && (
          <GuardrailsList guardrails={guardrails} onToggle={handleToggle} />
        )}
      </div>
    </>
  );
}

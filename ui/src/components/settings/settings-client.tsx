'use client';

import type { SettingCard } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState } from '@/components/admin-surface-state';
import { SettingsCardComponent } from './settings-card';

interface SettingsClientProps {
  cards: SettingCard[];
  unavailable: boolean;
}

export function SettingsClient({ cards, unavailable }: SettingsClientProps) {
  return (
    <>
      <PageHeader title="Settings" />
      <div className="content-area">
        {unavailable && <UnavailableState message="Settings are not yet available in this environment." />}
        {!unavailable && (
          <div className="space-y-4">
            {cards.map((card) => (
              <SettingsCardComponent key={card.id} card={card} />
            ))}
          </div>
        )}
      </div>
    </>
  );
}

import { fetchSettings } from '@/components/settings/settings-data';
import { SettingsClient } from '@/components/settings/settings-client';

export default async function SettingsPage() {
  const { cards, unavailable } = await fetchSettings();
  return <SettingsClient cards={cards} unavailable={unavailable} />;
}

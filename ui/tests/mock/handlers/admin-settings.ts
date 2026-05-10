import { Hono } from 'hono';
import { settingCards } from '../../fixtures/settings';
import type { SettingCard } from '../../../src/lib/admin-surface-contracts';

const displayPrefs: Record<string, string> = { theme: 'system', timezone: 'UTC' };

const app = new Hono();

app.get('/', (c) => {
  const cards: SettingCard[] = settingCards.map((card) => {
    if (card.id !== 'display') return card;
    return {
      ...card,
      fields: card.fields.map((f) => ({ ...f, value: displayPrefs[f.id] ?? f.value })),
    };
  });
  return c.json({ data: cards });
});

app.patch('/display', async (c) => {
  const body = await c.req.json<Record<string, string>>();
  Object.assign(displayPrefs, body);
  return c.json({ ok: true });
});

export default app;

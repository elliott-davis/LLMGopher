import type { SettingCard } from '../../src/lib/admin-surface-contracts';

export const settingCards: SettingCard[] = [
  {
    id: 'gateway-profile',
    title: 'Gateway Profile',
    description: 'Organization name and contact details.',
    availability: 'unavailable',
    fields: [
      { id: 'org_name', label: 'Organization name', value: 'LLMGopher Corp', input_type: 'text', read_only: true },
    ],
    save_capability: false,
  },
  {
    id: 'security',
    title: 'Security',
    description: 'Authentication and key rotation settings.',
    availability: 'unavailable',
    fields: [
      { id: 'api_key_rotation', label: 'Auto-rotate keys', value: 'disabled', input_type: 'select', read_only: true },
    ],
    save_capability: false,
  },
  {
    id: 'notifications',
    title: 'Notifications',
    description: 'Alert destinations for budget and rate-limit events.',
    availability: 'unavailable',
    fields: [
      { id: 'alert_email', label: 'Alert email', value: '', input_type: 'email', read_only: true },
    ],
    save_capability: false,
  },
  {
    id: 'display',
    title: 'Display',
    description: 'Local display preferences.',
    availability: 'editable',
    fields: [
      { id: 'theme', label: 'Theme', value: 'system', input_type: 'select', read_only: false },
      { id: 'timezone', label: 'Timezone', value: 'UTC', input_type: 'select', read_only: false },
    ],
    save_capability: true,
  },
];

'use client';

import { useState } from 'react';
import type { SettingCard, SettingField } from '@/lib/admin-surface-contracts';
import { redactValue } from '@/lib/redaction';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';

interface SettingsCardProps {
  card: SettingCard;
}

export function SettingsCardComponent({ card }: SettingsCardProps) {
  const [fieldValues, setFieldValues] = useState<Record<string, string>>(
    Object.fromEntries(card.fields.map((f) => [f.id, f.value]))
  );
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!card.save_capability) return;
    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      const res = await fetch(`/api/settings/${card.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(fieldValues),
      });
      if (res.ok) {
        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 3000);
      } else {
        setSaveError('Save failed. Please try again.');
      }
    } catch {
      setSaveError('Service unavailable.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card data-testid={`settings-card-${card.id}`}>
      <CardHeader>
        <CardTitle>{card.title}</CardTitle>
        <CardDescription>{card.description}</CardDescription>
      </CardHeader>
      <CardContent>
        {card.availability === 'unavailable' && (
          <div data-testid="settings-card-unavailable" className="text-sm text-muted-foreground p-3 border rounded bg-muted">
            This setting is not yet available. Production support will be enabled when the gateway configuration API is reconciled.
          </div>
        )}
        {card.availability !== 'unavailable' && (
          <div className="space-y-3">
            {card.fields.map((field) => (
              <FieldRow
                key={field.id}
                field={field}
                value={fieldValues[field.id] ?? field.value}
                readOnly={field.read_only || card.availability === 'read_only'}
                onChange={(v) => setFieldValues((prev) => ({ ...prev, [field.id]: v }))}
              />
            ))}
            {saveSuccess && (
              <p data-testid="settings-card-success" className="text-sm text-green-700">Saved successfully.</p>
            )}
            {saveError && (
              <p data-testid="settings-card-error" role="alert" className="text-sm text-red-700">{saveError}</p>
            )}
            {card.save_capability && (
              <button
                data-testid="settings-card-save"
                onClick={handleSave}
                disabled={saving}
                className="px-3 py-1.5 rounded text-sm bg-primary text-primary-foreground disabled:opacity-50"
              >
                {saving ? 'Saving…' : 'Save'}
              </button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function FieldRow({
  field,
  value,
  readOnly,
  onChange,
}: {
  field: SettingField;
  value: string;
  readOnly: boolean;
  onChange: (v: string) => void;
}) {
  const displayValue = field.id.toLowerCase().includes('key') || field.id.toLowerCase().includes('secret') || field.id.toLowerCase().includes('token')
    ? redactValue(value)
    : value;

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={`field-${field.id}`} className="text-xs font-medium text-muted-foreground">
        {field.label}
      </label>
      {readOnly ? (
        <p id={`field-${field.id}`} className="text-sm">{displayValue || '—'}</p>
      ) : (
        <input
          id={`field-${field.id}`}
          type={field.input_type === 'email' ? 'email' : 'text'}
          value={displayValue}
          onChange={(e) => onChange(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        />
      )}
      {field.validation_message && (
        <p role="alert" className="text-xs text-red-600">{field.validation_message}</p>
      )}
    </div>
  );
}

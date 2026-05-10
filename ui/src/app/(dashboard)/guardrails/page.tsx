import { fetchGuardrails } from '@/components/guardrails/guardrails-data';
import { GuardrailsClient } from '@/components/guardrails/guardrails-client';

export default async function GuardrailsPage() {
  const { guardrails, unavailable } = await fetchGuardrails();
  return <GuardrailsClient guardrails={guardrails} unavailable={unavailable} />;
}

'use client';

import { useState } from 'react';
import type { LogDetail } from '@/lib/admin-surface-contracts';
import { redactHeaders, redactPromptPreview, redactResponsePreview } from '@/lib/redaction';

type InspectorTab = 'trace' | 'prompt' | 'response' | 'headers';

interface RequestInspectorProps {
  detail: LogDetail;
  onClose: () => void;
}

export function RequestInspector({ detail, onClose }: RequestInspectorProps) {
  const [activeTab, setActiveTab] = useState<InspectorTab>('trace');
  const redactedHeaders = redactHeaders(detail.headers ?? {});

  return (
    <aside data-testid="request-inspector" className="fixed inset-y-0 right-0 w-[480px] bg-background border-l shadow-xl flex flex-col z-50">
      <div className="flex items-center justify-between px-4 py-3 border-b">
        <h2 className="text-sm font-medium">Request {detail.request_id}</h2>
        <button data-testid="inspector-close" onClick={onClose} aria-label="Close inspector">✕</button>
      </div>

      <div className="flex border-b px-4 gap-2">
        {(['trace', 'prompt', 'response', 'headers'] as const).map((tab) => (
          <button
            key={tab}
            data-testid={`inspector-tab-${tab}`}
            onClick={() => setActiveTab(tab)}
            className={`py-2 text-sm border-b-2 ${activeTab === tab ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'trace' && (
          <TraceTab trace={detail.trace ?? detail.provider_chain ?? []} />
        )}
        {activeTab === 'prompt' && (
          <pre className="text-xs whitespace-pre-wrap">{redactPromptPreview(detail.prompt_preview ?? '')}</pre>
        )}
        {activeTab === 'response' && (
          <pre className="text-xs whitespace-pre-wrap">{redactResponsePreview(detail.response_preview ?? '')}</pre>
        )}
        {activeTab === 'headers' && (
          <HeadersTab headers={redactedHeaders} />
        )}
      </div>
    </aside>
  );
}

function TraceTab({ trace }: { trace: LogDetail['provider_chain'] }) {
  return (
    <ol className="space-y-2">
      {trace.map((stage, i) => (
        <li
          key={stage.provider_id}
          data-testid={i === 0 ? 'timeline-stage-primary' : `timeline-stage-${i}`}
          data-failed={stage.status === 'failed' ? 'true' : undefined}
          className={`rounded p-3 border text-sm ${stage.status === 'failed' ? 'border-red-300 bg-red-50' : 'border-green-300 bg-green-50'}`}
        >
          <div className="font-medium">{stage.provider_id}</div>
          <div className="text-xs text-muted-foreground capitalize">{stage.status} · {stage.latency_ms}ms</div>
        </li>
      ))}
    </ol>
  );
}

function HeadersTab({ headers }: { headers: Record<string, string> }) {
  return (
    <dl className="space-y-1 text-xs">
      {Object.entries(headers).map(([k, v]) => (
        <div key={k} className="flex gap-2">
          <dt className="font-medium text-muted-foreground min-w-[140px]">{k}</dt>
          <dd>{v}</dd>
        </div>
      ))}
    </dl>
  );
}

const HTTP_REASON_PHRASES: Record<number, string> = {
  100: 'Continue',
  101: 'Switching Protocols',
  200: 'OK',
  201: 'Created',
  202: 'Accepted',
  204: 'No Content',
  301: 'Moved Permanently',
  302: 'Found',
  304: 'Not Modified',
  400: 'Bad Request',
  401: 'Unauthorized',
  403: 'Forbidden',
  404: 'Not Found',
  405: 'Method Not Allowed',
  408: 'Request Timeout',
  409: 'Conflict',
  410: 'Gone',
  422: 'Unprocessable Entity',
  429: 'Too Many Requests',
  500: 'Internal Server Error',
  501: 'Not Implemented',
  502: 'Bad Gateway',
  503: 'Service Unavailable',
  504: 'Gateway Timeout',
};

export function formatStatusLabel(statusCode: number): string {
  const phrase = HTTP_REASON_PHRASES[statusCode];
  return phrase ? `${statusCode} ${phrase}` : String(statusCode);
}

export function formatLatency(ms: number): string {
  if (ms === 0) return '<1ms';
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export function formatCurrency(usd: number): string {
  return `$${usd.toFixed(4)}`;
}

export function formatTokenCount(tokens: number): string {
  if (tokens < 1000) return String(tokens);
  if (tokens < 1_000_000) return `${(tokens / 1000).toFixed(1)}k`;
  return `${(tokens / 1_000_000).toFixed(1)}M`;
}

export function formatPercentage(ratio: number): string {
  return `${Math.round(ratio * 100)}%`;
}

export function utilizationState(
  ratio: number,
  threshold: number
): 'ok' | 'near_cap' | 'over_cap' {
  if (ratio >= 1.0) return 'over_cap';
  if (ratio >= threshold) return 'near_cap';
  return 'ok';
}

export function unavailableCopy(surface: string): string {
  return `${surface} is not yet available in this environment.`;
}

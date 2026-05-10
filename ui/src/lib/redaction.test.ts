import { describe, expect, it } from 'vitest';

import {
  redactErrorSummary,
  redactHeaders,
  redactPromptPreview,
  redactValue,
  truncatePreview,
} from './redaction';

describe('redactHeaders', () => {
  it('redacts authorization header', () => {
    const result = redactHeaders({ authorization: 'Bearer sk-abc123' });
    expect(result['authorization']).toBe('[REDACTED]');
  });

  it('redacts cookie header', () => {
    const result = redactHeaders({ cookie: 'session=abc' });
    expect(result['cookie']).toBe('[REDACTED]');
  });

  it('redacts x-api-key header', () => {
    const result = redactHeaders({ 'x-api-key': 'sk-test-key-1' });
    expect(result['x-api-key']).toBe('[REDACTED]');
  });

  it('redacts x-auth-token header', () => {
    const result = redactHeaders({ 'x-auth-token': 'some-token-value' });
    expect(result['x-auth-token']).toBe('[REDACTED]');
  });

  it('preserves safe headers', () => {
    const result = redactHeaders({
      'content-type': 'application/json',
      'content-length': '42',
      'x-request-id': 'req-001',
    });
    expect(result['content-type']).toBe('application/json');
    expect(result['content-length']).toBe('42');
    expect(result['x-request-id']).toBe('req-001');
  });

  it('redacts Bearer value even when partially masked', () => {
    const result = redactHeaders({ authorization: 'Bearer ****' });
    expect(result['authorization']).toBe('[REDACTED]');
  });
});

describe('redactValue', () => {
  it('returns [REDACTED] for bearer tokens', () => {
    expect(redactValue('Bearer sk-abc123')).toBe('[REDACTED]');
  });

  it('returns [REDACTED] for sk- prefixed values', () => {
    expect(redactValue('sk-test-key-1:key-001')).toBe('[REDACTED]');
  });
});

describe('truncatePreview', () => {
  it('truncates strings longer than 120 chars', () => {
    const long = 'a'.repeat(200);
    const result = truncatePreview(long);
    expect(result).toBe('a'.repeat(120) + '…');
  });

  it('returns strings at or under 120 chars unchanged', () => {
    const short = 'Hello world';
    expect(truncatePreview(short)).toBe(short);
    const exact = 'a'.repeat(120);
    expect(truncatePreview(exact)).toBe(exact);
  });
});

describe('redactPromptPreview', () => {
  it('truncates long prompts', () => {
    const long = 'Tell me about '.repeat(20);
    const result = redactPromptPreview(long);
    expect(result.length).toBeLessThanOrEqual(121);
    expect(result.endsWith('…')).toBe(true);
  });
});

describe('redactErrorSummary', () => {
  it('redacts error messages containing key material', () => {
    expect(redactErrorSummary('Invalid api key provided')).toBe('[REDACTED]');
    expect(redactErrorSummary('Bearer token rejected')).toBe('[REDACTED]');
    expect(redactErrorSummary('sk-abc123 is unauthorized')).toBe('[REDACTED]');
  });

  it('truncates safe error messages', () => {
    const safe = 'Connection refused by upstream provider';
    expect(redactErrorSummary(safe)).toBe(safe);
  });
});
